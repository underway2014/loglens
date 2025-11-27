package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// 命令行参数
var (
	filePath        string
	unescapeFlag    bool
	helpFlag        bool
	interactiveFlag bool
	showLineNumber  bool
	keepOneLine     bool
	trimSpace       bool
)

func init() {
	flag.StringVar(&filePath, "f", "", "日志文件路径")
	flag.StringVar(&filePath, "file", "", "日志文件路径")
	flag.BoolVar(&unescapeFlag, "u", false, "是否替换转义符（如 \\n, \\t, \\r 等）")
	flag.BoolVar(&unescapeFlag, "unescape", false, "是否替换转义符（如 \\n, \\t, \\r 等）")
	flag.BoolVar(&interactiveFlag, "i", false, "交互式分页查看模式（默认已启用，此选项保留以兼容）")
	flag.BoolVar(&interactiveFlag, "interactive", false, "交互式分页查看模式（默认已启用，此选项保留以兼容）")
	flag.BoolVar(&showLineNumber, "l", false, "显示行号")
	flag.BoolVar(&showLineNumber, "line-number", false, "显示行号")
	flag.BoolVar(&keepOneLine, "k", false, "替换转义符时保持每条日志在一行（将\\n替换为空格）")
	flag.BoolVar(&keepOneLine, "keep-one-line", false, "替换转义符时保持每条日志在一行（将\\n替换为空格）")
	flag.BoolVar(&trimSpace, "t", false, "修剪每行开头和结尾的空白字符")
	flag.BoolVar(&trimSpace, "trim", false, "修剪每行开头和结尾的空白字符")
	flag.BoolVar(&helpFlag, "h", false, "显示帮助信息")
	flag.BoolVar(&helpFlag, "help", false, "显示帮助信息")
}

func main() {
	flag.Parse()

	// 显示帮助信息
	if helpFlag {
		showHelp()
		return
	}

	// 获取文件路径：优先使用 -f 参数，其次使用位置参数
	if filePath == "" && flag.NArg() > 0 {
		filePath = flag.Arg(0)
	}

	// 如果没有指定文件，从标准输入读取
	if filePath == "" {
		if err := processStream(os.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开文件 %s: %v\n", filePath, err)
		os.Exit(1)
	}
	file.Close() // 立即关闭，交互模式会重新打开

	// 检查是否是管道输出或非交互式终端
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		// 输出被重定向，使用非交互模式
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无法打开文件 %s: %v\n", filePath, err)
			os.Exit(1)
		}
		defer file.Close()
		if err := processStream(file); err != nil {
			fmt.Fprintf(os.Stderr, "读取文件错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 默认使用交互式模式
	if err := interactiveMode(filePath); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

// processStream 处理输入流并输出
func processStream(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if trimSpace {
			line = strings.TrimSpace(line)
		}
		if unescapeFlag {
			line = unescapeString(line)
		}
		if showLineNumber {
			// 使用青色（Cyan）显示行号
			fmt.Printf("\033[36m%6d\033[0m  %s\n", lineNum, line)
		} else {
			fmt.Println(line)
		}
		lineNum++
	}
	return scanner.Err()
}

// unescapeString 替换字符串中的转义符
func unescapeString(s string) string {
	if keepOneLine {
		// 保持在一行：将 \n 替换为空格
		replacer := strings.NewReplacer(
			"\\t", "\t",
			"\\r", "",
			"\\n", " ", // 替换为空格
			"\\\\", "\\",
			"\\\"", "\"",
			"\\'", "'",
		)
		return replacer.Replace(s)
	} else {
		// 正常替换：\n 替换为真正的换行符
		replacer := strings.NewReplacer(
			"\\n", "\n",
			"\\t", "\t",
			"\\r", "\r",
			"\\\\", "\\",
			"\\\"", "\"",
			"\\'", "'",
		)
		return replacer.Replace(s)
	}
}

// isJSONLine 检查一行是否是 JSON 格式
func isJSONLine(line string) bool {
	var js interface{}
	return json.Unmarshal([]byte(line), &js) == nil
}

// formatJSONLine 格式化 JSON 行为美化的多行显示
func formatJSONLine(line string) (string, error) {
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(line), &jsonObj); err != nil {
		return line, err
	}

	formatted, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return line, err
	}

	return string(formatted), nil
}

// showFormattedJSON 在独立页面显示格式化的 JSON
func showFormattedJSON(filePath string, lineIndex []int64, lineNum int, width int) error {
	// 读取指定行的内容
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 定位到指定行
	if lineNum >= len(lineIndex) {
		return fmt.Errorf("行号超出范围")
	}

	_, err = file.Seek(lineIndex[lineNum], 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return fmt.Errorf("无法读取行内容")
	}

	rawLine := scanner.Text()
	if trimSpace {
		rawLine = strings.TrimSpace(rawLine)
	}
	if unescapeFlag {
		rawLine = unescapeString(rawLine)
	}

	// 检查是否是 JSON
	if !isJSONLine(rawLine) {
		// 清屏并显示错误信息
		fmt.Print("\033[2J\033[H")
		fmt.Printf("第 %d 行不是有效的 JSON 格式\r\n\r\n", lineNum+1)
		fmt.Printf("原始内容：\r\n%s\r\n\r\n", rawLine)
		fmt.Print("按任意键返回...")

		// 等待用户按键
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		return nil
	}

	// 格式化 JSON
	formatted, err := formatJSONLine(rawLine)
	if err != nil {
		// 清屏并显示错误信息
		fmt.Print("\033[2J\033[H")
		fmt.Printf("JSON 格式化失败: %v\r\n\r\n", err)
		fmt.Printf("原始内容：\r\n%s\r\n\r\n", rawLine)
		fmt.Print("按任意键返回...")

		// 等待用户按键
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		return nil
	}

	// 清屏并显示格式化的 JSON
	fmt.Print("\033[2J\033[H")
	fmt.Printf("\033[32m=== 第 %d 行 JSON 格式化 ===\033[0m\r\n\r\n", lineNum+1)

	// 按行输出格式化的 JSON，确保使用 \r\n
	lines := strings.Split(formatted, "\n")
	for _, line := range lines {
		fmt.Print(line + "\r\n")
	}

	fmt.Print("\r\n\033[90m按任意键返回...\033[0m")

	// 等待用户按键
	buf := make([]byte, 1)
	os.Stdin.Read(buf)

	return nil
}

// interactiveMode 交互式分页查看模式
func interactiveMode(filePath string) error {
	// 在交互模式下，如果启用了 unescape，自动开启 keepOneLine 模式
	// 避免 JSON 中的 \n 被替换成真正的换行符导致行索引错乱
	originalKeepOneLine := keepOneLine
	if unescapeFlag {
		keepOneLine = true
	}
	defer func() {
		keepOneLine = originalKeepOneLine
	}()

	// 按需读取文件行索引
	lineIndex, err := buildLineIndex(filePath)
	if err != nil {
		return err
	}

	// lineIndex 包含每行的起始位置
	// 例如：3行文件会有 [0, pos1, pos2]，长度为3
	// 但如果文件末尾有换行符，会多一个位置 [0, pos1, pos2, pos3]，长度为4
	// 实际行数应该是最后一个位置之前的元素个数
	// 如果最后一个元素指向文件末尾且那里没有内容，说明最后一行后面有换行符
	totalLines := len(lineIndex)
	// 如果 lineIndex 最后一个元素等于文件大小，说明最后一行有换行符，实际行数要减1
	fileInfo, err := os.Stat(filePath)
	if err == nil && totalLines > 0 && lineIndex[totalLines-1] == fileInfo.Size() {
		totalLines--
	}

	if totalLines == 0 {
		fmt.Println("文件为空")
		return nil
	}

	// 获取终端大小
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// 如果无法获取终端大小，使用默认值
		height = 20
		width = 80
	}

	// 计算实际可用的显示行数
	// 使用整个终端高度减去1行缓冲
	viewHeight := height - 1
	if viewHeight < 3 {
		viewHeight = 3 // 最少显示3行，即使窗口很小
	}

	// 保存原始终端状态
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("无法进入原始终端模式: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// 监听窗口大小变化信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	// 用于标记是否需要重新显示
	needRedisplay := false

	// 在后台监听窗口大小变化
	go func() {
		for range sigChan {
			needRedisplay = true
		}
	}()

	currentLine := 0
	lastDisplayedLine := 0   // 记录上次显示的最后一行
	searchPattern := ""      // 搜索模式
	searchMatches := []int{} // 搜索结果（匹配的行号）
	currentMatchIndex := -1  // 当前匹配的索引

	// 显示第一页
	lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
	if err != nil {
		return err
	}
	lastDisplayedLine = lastLine

	// 主循环
	buf := make([]byte, 1)
	commandBuf := []byte{}
	for {
		// 检查是否需要因为窗口大小变化而重新显示
		if needRedisplay {
			needRedisplay = false
			// 重新获取终端大小
			newWidth, newHeight, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				width = newWidth
				height = newHeight
				viewHeight = height - 1
				if viewHeight < 3 {
					viewHeight = 3
				}
			}
			// 重新显示当前页
			lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
			if err == nil {
				lastDisplayedLine = lastLine
			}
			continue
		}

		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}

		ch := buf[0]

		// 处理命令模式
		if len(commandBuf) > 0 {
			// 已经在命令模式中，处理命令输入
			if ch == '\r' || ch == '\n' {
				// 执行命令
				cmdType := commandBuf[0]
				cmd := string(commandBuf[1:])

				if cmdType == ':' {
					// 检查是否是格式化命令 :f<行号>
					if strings.HasPrefix(cmd, "f") {
						// 格式化指定行的 JSON
						lineNumStr := strings.TrimPrefix(cmd, "f")
						if lineNumStr == "" {
							// 如果没有指定行号，使用当前行
							err := showFormattedJSON(filePath, lineIndex, currentLine, width)
							if err != nil {
								// 如果出错，仅记录错误但不退出程序
							}
						} else if lineNum, err := strconv.Atoi(lineNumStr); err == nil {
							if lineNum > 0 && lineNum <= totalLines {
								// 格式化指定行（转为 0 基索引）
								err := showFormattedJSON(filePath, lineIndex, lineNum-1, width)
								if err != nil {
									// 如果出错，仅记录错误但不退出程序
								}
							}
						}
					} else {
						// 普通的跳转命令
						if lineNum, err := strconv.Atoi(cmd); err == nil {
							if lineNum > 0 && lineNum <= totalLines {
								currentLine = lineNum - 1
							}
						}
					}
				} else if cmdType == '/' {
					// 搜索
					if cmd != "" {
						searchPattern = cmd
						// 执行搜索
						searchMatches = searchInFile(filePath, totalLines, searchPattern)
						if len(searchMatches) > 0 {
							// 跳转到第一个匹配
							currentMatchIndex = 0
							currentLine = searchMatches[0]
						}
					}
				}

				commandBuf = []byte{}
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
				continue
			}
			if ch == 27 { // ESC - 取消命令
				commandBuf = []byte{}
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
				continue
			}
			if ch == 127 || ch == 8 { // Backspace - 删除字符
				if len(commandBuf) > 1 {
					commandBuf = commandBuf[:len(commandBuf)-1]
					fmt.Print("\b \b")
				}
				continue
			}
			// 添加字符到命令缓冲区（包括冒号等特殊字符）
			commandBuf = append(commandBuf, ch)
			fmt.Printf("%c", ch)
			continue
		} else if ch == ':' || ch == '/' {
			// 开启命令模式
			commandBuf = []byte{ch}
			fmt.Printf("\r\n%c", ch)
			continue
		}

		switch ch {
		case 'q', 'Q':
			fmt.Print("\r\n")
			return nil
		case 'n': // 下一个搜索匹配
			if len(searchMatches) > 0 && currentMatchIndex >= 0 {
				currentMatchIndex++
				if currentMatchIndex >= len(searchMatches) {
					currentMatchIndex = 0 // 循环到第一个
				}
				currentLine = searchMatches[currentMatchIndex]
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case 'N': // 上一个搜索匹配
			if len(searchMatches) > 0 && currentMatchIndex >= 0 {
				currentMatchIndex--
				if currentMatchIndex < 0 {
					currentMatchIndex = len(searchMatches) - 1 // 循环到最后一个
				}
				currentLine = searchMatches[currentMatchIndex]
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case 6: // Ctrl+F - 前翻页（下一页）
			// 翻页时保持连续：上一页的最后一行成为新页的第一行
			if lastDisplayedLine < totalLines-1 {
				// 如果没有显示到新的内容（当前行太长），强制往前跳一行
				if lastDisplayedLine == currentLine {
					currentLine++
					if currentLine >= totalLines {
						currentLine = totalLines - 1
					}
				} else {
					currentLine = lastDisplayedLine
				}
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case 2: // Ctrl+B - 后翻页（上一页）
			// vim 风格：当前页的第一行成为新页的最后一行（或最后几行之一）
			// 策略：往前找，找到一个起始位置，使得显示后最后一行接近 currentLine
			if currentLine > 0 {
				// 二分查找：找到合适的起始位置
				// 初始范围：[0, currentLine)
				left := 0
				right := currentLine
				bestStart := 0

				// 最多尝试15次二分查找
				for attempt := 0; attempt < 15 && left < right; attempt++ {
					mid := (left + right) / 2
					if mid == bestStart {
						// 避免死循环
						break
					}
					testLast, _ := displayPage(filePath, lineIndex, mid, totalLines, viewHeight, width, searchPattern)

					if testLast < currentLine {
						// 显示的最后一行还没到 currentLine，起始位置太靠后了
						left = mid + 1
						bestStart = mid
					} else if testLast > currentLine {
						// 显示的最后一行超过了 currentLine，起始位置太靠前了
						right = mid
					} else {
						// 正好！
						bestStart = mid
						break
					}
				}

				currentLine = bestStart
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case 'j', 'J': // j - 下一行
			if lastDisplayedLine < totalLines-1 {
				currentLine++
				if currentLine >= totalLines {
					currentLine = totalLines - 1
				}
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case 'k', 'K': // k - 上一行
			if currentLine > 0 {
				currentLine--
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case '\n', '\r': // Enter - 下一行
			if lastDisplayedLine < totalLines-1 {
				currentLine++
				if currentLine >= totalLines {
					currentLine = totalLines - 1
				}
				lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
				if err != nil {
					return err
				}
				lastDisplayedLine = lastLine
			}
		case 'g': // 第一页
			currentLine = 0
			lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
			if err != nil {
				return err
			}
			lastDisplayedLine = lastLine
		case 'G': // 最后一行
			// 跳转到最后一行
			currentLine = totalLines - 1
			if currentLine < 0 {
				currentLine = 0
			}
			lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
			if err != nil {
				return err
			}
			lastDisplayedLine = lastLine
		case 'f', 'F': // f - 格式化当前行的 JSON
			// 显示 JSON 格式化页面
			err := showFormattedJSON(filePath, lineIndex, currentLine, width)
			if err != nil {
				// 如果出错，仅记录错误但不退出程序
				// 可以在这里显示错误信息
			}
			// 返回后重新显示当前页
			lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
			if err != nil {
				return err
			}
			lastDisplayedLine = lastLine
		case 27: // ESC sequence for arrow keys
			// 读取下两个字节
			next := make([]byte, 2)
			os.Stdin.Read(next)
			if next[0] == '[' {
				if next[1] == 'A' { // 上箭头 - 上一行
					if currentLine > 0 {
						currentLine--
						lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
						if err != nil {
							return err
						}
						lastDisplayedLine = lastLine
					}
				} else if next[1] == 'B' { // 下箭头 - 下一行
					if currentLine+1 < totalLines {
						currentLine++
						lastLine, err := displayPage(filePath, lineIndex, currentLine, totalLines, viewHeight, width, searchPattern)
						if err != nil {
							return err
						}
						lastDisplayedLine = lastLine
					}
				}
			}
		}
	}

	fmt.Print("\r\n")
	return nil
}

// buildLineIndex 构建文件行索引（记录每行的起始位置）
func buildLineIndex(filePath string) ([]int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lineIndex []int64
	lineIndex = append(lineIndex, 0) // 第一行从位置 0 开始

	reader := bufio.NewReader(file)
	var pos int64 = 0

	for {
		line, err := reader.ReadBytes('\n')
		lineLen := int64(len(line))

		if err == io.EOF {
			// 如果文件最后一行没有换行符，这也算一行
			if lineLen > 0 {
				// 已经记录了这一行的起始位置，不需要再添加
			}
			break
		}
		if err != nil {
			return nil, err
		}

		pos += lineLen
		// 记录下一行的起始位置
		lineIndex = append(lineIndex, pos)
	}

	// 返回每行的起始位置数组
	// 如果文件有N行且最后一行有换行符，会有N+1个位置（最后一个指向EOF）
	// 如果文件有N行但最后一行没有换行符，会有N个位置
	return lineIndex, nil
}

// displayPage 显示指定页的内容，返回实际显示的最后一行的索引
// 返回值：lastDisplayedLine - 实际显示的最后一行索引
func displayPage(filePath string, lineIndex []int64, startLine, totalLines, viewHeight, termWidth int, searchPattern string) (int, error) {
	// 清屏
	fmt.Print("\033[2J\033[H")

	// 计算结束行
	endLine := startLine + viewHeight*2 // 多读一些行，以防有的行很短
	if endLine > totalLines {
		endLine = totalLines
	}

	// 读取并显示指定范围的行
	file, err := os.Open(filePath)
	if err != nil {
		return startLine, err
	}
	defer file.Close()

	// 定位到起始行（即使 startLine=0 也要明确 seek 到开头）
	_, err = file.Seek(lineIndex[startLine], 0)
	if err != nil {
		return startLine, err
	}

	scanner := bufio.NewScanner(file)

	// 计算实际使用的终端行数
	screenLinesUsed := 0
	lastDisplayedLine := startLine - 1 // 记录实际显示的最后一行

	for i := startLine; i < endLine && scanner.Scan(); i++ {
		line := scanner.Text()
		if trimSpace {
			line = strings.TrimSpace(line)
		}
		if unescapeFlag {
			line = unescapeString(line)
		}

		// 如果有搜索模式，高亮匹配的字符串
		if searchPattern != "" {
			line = highlightMatches(line, searchPattern)
		}

		// 计算这一行显示时会占用多少终端行
		// 行号占用的宽度（如果显示行号）
		linePrefix := ""
		if showLineNumber {
			linePrefix = fmt.Sprintf("\033[36m%6d\033[0m  ", i+1)
		}

		// 计算内容宽度（考虑行号前缀的显示宽度，ANSI颜色码不占宽度）
		prefixWidth := 8 // "  1234  " 的可见宽度
		if !showLineNumber {
			prefixWidth = 0
		}

		availableWidth := termWidth - prefixWidth
		if availableWidth < 10 {
			availableWidth = 10 // 最小宽度
		}

		// 计算这一行会占用多少屏幕行（向上取整）
		lineLength := len(line)
		linesNeeded := (lineLength + availableWidth - 1) / availableWidth
		if linesNeeded == 0 {
			linesNeeded = 1 // 空行至少占1行
		}

		// 检查是否超出屏幕
		// 如果剩余空间太少（小于3行），且这一行需要很多行，就不显示
		// 但如果这是第一行，无论多长都要显示（避免空白屏幕）
		remainingLines := viewHeight - screenLinesUsed
		if i > startLine && remainingLines < 3 && linesNeeded > remainingLines {
			break // 不显示这一行，避免超出屏幕太多
		}

		// 如果这一行会超出屏幕，但剩余空间还有几行，就部分显示
		if screenLinesUsed+linesNeeded > viewHeight && remainingLines > 0 {
			// 截断显示：只显示能放下的部分
			if remainingLines > 0 {
				maxChars := remainingLines * availableWidth
				// 注意：line 可能包含 ANSI 颜色码，直接截断可能不准确
				// 但为了简化，我们先检查长度
				if lineLength > maxChars && maxChars > 0 {
					// 安全截断：确保不超出字符串范围
					if maxChars < len(line) {
						line = line[:maxChars] + "..."
					}
					linesNeeded = remainingLines
				}
			}
		}

		// 显示这一行
		if showLineNumber {
			fmt.Printf("%s%s\r\n", linePrefix, line)
		} else {
			fmt.Printf("%s\r\n", line)
		}

		screenLinesUsed += linesNeeded
		lastDisplayedLine = i // 更新实际显示的最后一行
	}

	if err := scanner.Err(); err != nil {
		return lastDisplayedLine, err
	}

	// 不显示状态栏

	return lastDisplayedLine, nil
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("LogLens (lg) - 日志查看工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  lg [选项] [文件路径]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -l, --line-number        显示行号")
	fmt.Println("  -u, --unescape           替换转义符（\\n, \\t, \\r 等）")
	fmt.Println("  -k, --keep-one-line      配合 -u 使用，保持每条日志在一行（\\n替换为空格）")
	fmt.Println("  -t, --trim               修剪每行开头和结尾的空白字符")
	fmt.Println("  -h, --help               显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  lg app.log                        # 交互式查看 app.log（直接传文件路径）")
	fmt.Println("  lg app.log -l                     # 交互式查看并显示行号")
	fmt.Println("  lg app.log -u -k                  # 交互式查看并替换转义符（保持单行）")
	fmt.Println("  cat app.log | lg -u               # 从管道读取并替换转义符")
	fmt.Println("  lg app.log > output.txt           # 输出重定向（自动使用非交互模式）")
	fmt.Println()
	fmt.Println("交互式模式命令:")
	fmt.Println("  Ctrl+F          下一页")
	fmt.Println("  Ctrl+B          上一页")
	fmt.Println("  j/↓             下一行")
	fmt.Println("  k/↑             上一行")
	fmt.Println("  Enter           下一行")
	fmt.Println("  g               跳转到第一页")
	fmt.Println("  G               跳转到最后一行")
	fmt.Println("  :<行号>         跳转到指定行（例如 :100 跳转到第100行）")
	fmt.Println("  :f              格式化当前行为 JSON")
	fmt.Println("  :f<行号>        格式化指定行为 JSON（例如 :f5 格式化第5行）")
	fmt.Println("  /<模式>         搜索（简单字符串搜索，不区分大小写）")
	fmt.Println("  n               下一个搜索匹配")
	fmt.Println("  N               上一个搜索匹配")
	fmt.Println("  f               格式化当前行为 JSON（快捷键）")
	fmt.Println("  q               退出")
	fmt.Println()
}

// searchInFile 在文件中搜索匹配的行，返回匹配的行号列表
// 使用简单的字符串搜索（不区分大小写）
func searchInFile(filePath string, totalLines int, pattern string) []int {
	var matches []int

	if pattern == "" {
		return matches
	}

	file, err := os.Open(filePath)
	if err != nil {
		return matches
	}
	defer file.Close()

	// 转换为小写用于不区分大小写搜索
	lowerPattern := strings.ToLower(pattern)

	scanner := bufio.NewScanner(file)
	for i := 0; i < totalLines && scanner.Scan(); i++ {
		line := scanner.Text()

		// 应用与显示相同的处理
		if trimSpace {
			line = strings.TrimSpace(line)
		}
		if unescapeFlag {
			line = unescapeString(line)
		}

		// 简单字符串搜索（不区分大小写）
		if strings.Contains(strings.ToLower(line), lowerPattern) {
			matches = append(matches, i)
		}
	}

	return matches
}

// highlightMatches 高亮匹配的字符串（使用黄色背景 + 黑色文字）
// 使用简单的字符串搜索（不区分大小写）
func highlightMatches(line string, pattern string) string {
	if pattern == "" {
		return line
	}

	// 普通字符串搜索（不区分大小写）
	// 需要找到原始大小写并高亮
	lowerLine := strings.ToLower(line)
	lowerPattern := strings.ToLower(pattern)

	result := ""
	lastIdx := 0

	for {
		idx := strings.Index(lowerLine[lastIdx:], lowerPattern)
		if idx == -1 {
			break
		}

		actualIdx := lastIdx + idx

		// 添加匹配前的部分
		result += line[lastIdx:actualIdx]

		// 添加高亮的匹配部分（保持原始大小写）
		matchedText := line[actualIdx : actualIdx+len(pattern)]
		result += "\033[43;30m" + matchedText + "\033[0m"

		lastIdx = actualIdx + len(pattern)
	}

	// 添加剩余部分
	result += line[lastIdx:]

	return result
}
