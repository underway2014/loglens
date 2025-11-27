# LogLens (lg) - 日志查看工具

一个用 Go 语言实现的交互式命令行日志查看工具，类似 `less`，但专为日志文件优化，特别适合查看 JSON 日志。

## ✨ 功能特性

- ✅ **交互式分页** - 默认交互式分页模式，按需读取文件内容，不会一次性加载所有日志
- ✅ **vim 风格导航** - 支持 `j`/`k`/`g`/`G`/`Ctrl+F`/`Ctrl+B` 等 vim 快捷键
- ✅ **搜索功能** - `/` 搜索关键词，`n`/`N` 在匹配间导航，黄色高亮显示
- ✅ **JSON 格式化** - 按 `f` 键可将当前行格式化为美化的 JSON（适合 JSON 日志）
- ✅ **行号显示** - 可选显示行号，方便定位
- ✅ **转义符替换** - 可选的转义符替换（`\n`, `\t`, `\r`, `\"`, `\'`, `\\`）
- ✅ **快速跳转** - 支持跳转到指定行、首页、尾页
- ✅ **支持管道输入** - 可以从标准输入读取
- ✅ **内存优化** - 按需读取，不会将整个文件加载到内存

## 安装

### 二进制发布版（推荐）

从 [GitHub Releases](https://github.com/underway2014/loglens/releases) 下载预编译的二进制文件,重命名为lg放到系统对应的可执行目录

## 使用方法

### 基本用法

```bash
# 查看帮助信息
lg -h

# 交互式查看日志（默认模式）
lg app.log

# 显示行号
lg app.log -l

# 替换转义符（适合 JSON 日志）
lg app.log -u -k

# 全功能模式：行号 + 转义符替换
lg app.log -l -u -k

# 从管道读取并替换转义符
cat app.log | lg -u

# 输出重定向（自动使用非交互模式）
lg app.log > output.txt

# 配合其他命令
tail -f app.log | lg -u       # 实时查看日志
tail -n 100 app.log | lg -u   # 查看最后 100 行
```

### 命令行参数

| 参数 | 简写 | 说明 |
|------|------|------|
| `--line-number` | `-l` | 显示行号 |
| `--unescape` | `-u` | 替换转义符（如 `\n`, `\t`, `\r` 等） |
| `--keep-one-line` | `-k` | 配合 -u 使用，保持每条日志在一行（`\n`替换为空格） |
| `--trim` | `-t` | 修剪每行开头和结尾的空白字符 |
| `--help` | `-h` | 显示帮助信息 |

### 交互式模式命令

默认启用交互式模式，支持以下快捷键：

| 按键 | 功能 |
|------|------|
| `Ctrl+F` | 下一页 |
| `Ctrl+B` | 上一页 |
| `Enter` / `j` / `↓` | 下一行 |
| `k` / `↑` | 上一行 |
| `g` | 跳转到第一页 |
| `G` | 跳转到最后一行 |
| `:<行号>` | **跳转到指定行**（例如 `:100` 跳转到第 100 行） |
| `:f` | **格式化当前行**（将当前行格式化为 JSON） |
| `:f<行号>` | **格式化指定行**（例如 `:f5` 格式化第 5 行） |
| `/<模式>` | **搜索**（简单字符串搜索，不区分大小写） |
| `n` | 下一个搜索匹配 |
| `N` | 上一个搜索匹配 |
| `f` | **JSON 格式化**（格式化当前行为美化的 JSON） |
| `q` | 退出交互模式 |

## 🎯 核心功能详解

### 1. 交互式分页查看

**特点：按需读取，不会一次性加载整个文件到内存**

```bash
# 基础用法
lg large.log

# 带行号
lg large.log -l

# 处理 JSON 日志
lg app.log -u -k
```

在交互模式下：
- 按 `Ctrl+F` 查看下一页
- 按 `Ctrl+B` 查看上一页
- 按 `j` 或 `Enter` 或 `↓` 下移一行
- 按 `k` 或 `↑` 上移一行
- 按 `g` 跳转到文件开头
- 按 `G` 跳转到文件末尾
- 输入 `:100` 然后按 `Enter` 跳转到第 100 行
- 输入 `/error` 然后按 `Enter` 搜索 "error"，按 `n`/`N` 在匹配间导航
- 按 `f` 格式化当前行的 JSON（如果是有效 JSON）
- 按 `q` 退出

### 2. 搜索功能

在交互模式下，支持实时搜索和高亮显示：

1. 按 `/` 键进入搜索模式
2. 输入搜索关键词（不区分大小写）
3. 按 `Enter` 开始搜索
4. 按 `n` 跳转到下一个匹配
5. 按 `N` 跳转到上一个匹配

示例：
```
/error    # 搜索 "error"
/404      # 搜索 "404"
/success  # 搜索 "success"
```

搜索结果会以黄色背景高亮显示。

### 3. JSON 格式化功能

在交互模式下，可以对单行 JSON 数据进行格式化显示：

**方法一：快捷键**
1. 将光标移动到包含 JSON 的行
2. 按 `f` 键格式化 JSON

**方法二：命令模式**
1. 按 `:` 进入命令模式
2. 输入 `f` 格式化当前行
3. 或输入 `f<行号>` 格式化指定行（例如 `f5` 格式化第 5 行）
4. 按 `Enter` 确认

**示例**：

原始 JSON 行：
```
{"timestamp":"2025-11-27T10:00:01","level":"INFO","message":"User login","user_id":12345,"metadata":{"browser":"Chrome"}}
```

按 `f` 键后显示：
```json
{
  "timestamp": "2025-11-27T10:00:01",
  "level": "INFO",
  "message": "User login",
  "user_id": 12345,
  "metadata": {
    "browser": "Chrome"
  }
}
```

注意：
- 只有当前行是有效的 JSON 格式时才能格式化
- 非 JSON 行会显示错误信息和原始内容
- 支持复杂嵌套的 JSON 结构

### 4. 跳转到指定行

在交互模式下，可以快速跳转到任意行：

1. 按 `:` 键进入命令模式
2. 输入行号，例如 `50`
3. 按 `Enter` 确认跳转
4. 按 `ESC` 可以取消输入

示例：
```
:50    # 跳转到第 50 行
:1000  # 跳转到第 1000 行
:1     # 跳转到第 1 行（等同于按 g）
```

### 5. 显示行号

```bash
# 显示行号
lg app.log -l
```

行号格式为右对齐 6 位数字，方便阅读。

### 6. 转义符替换示例

**原始日志内容（包含转义符）：**
```
2025-11-26 10:00:03 INFO User login: username=john\tIP=192.168.1.100
2025-11-26 10:00:04 WARN Connection timeout\nRetrying...
2025-11-26 10:00:06 INFO Request: {\"status\":\"success\"}
```

**不使用 `-u` 参数：**
```bash
lg test.log
```
输出（保留转义符）：
```
2025-11-26 10:00:03 INFO User login: username=john\tIP=192.168.1.100
2025-11-26 10:00:04 WARN Connection timeout\nRetrying...
2025-11-26 10:00:06 INFO Request: {\"status\":\"success\"}
```

**使用 `-u` 参数：**
```bash
lg test.log -u
```
输出（替换转义符）：
```
2025-11-26 10:00:03 INFO User login: username=john	IP=192.168.1.100
2025-11-26 10:00:04 WARN Connection timeout
Retrying...
2025-11-26 10:00:06 INFO Request: {"status":"success"}
```

**使用 `-u -k` 参数（保持单行）：**
```bash
lg test.log -u -k
```
输出（替换转义符但保持单行）：
```
2025-11-26 10:00:03 INFO User login: username=john	IP=192.168.1.100
2025-11-26 10:00:04 WARN Connection timeout Retrying...
2025-11-26 10:00:06 INFO Request: {"status":"success"}
```

## 支持的转义符

| 转义符 | 替换为 | 说明 |
|--------|--------|------|
| `\n` | 换行符 | 换行 |
| `\t` | 制表符 | Tab 键 |
| `\r` | 回车符 | 回车 |
| `\\` | `\` | 反斜杠 |
| `\"` | `"` | 双引号 |
| `\'` | `'` | 单引号 |

## 使用场景

1. **查看包含转义符的 JSON 日志**
   ```bash
   lg app.log -u -k
   ```

2. **搜索错误日志**
   ```bash
   lg app.log      # 然后按 / 输入 error 搜索
   ```

3. **格式化 JSON 日志**
   ```bash
   lg app.log -u           # 进入交互模式，按 f 键格式化当前行
   # 或者使用命令：:f 格式化当前行，:f5 格式化第5行
   ```

4. **实时监控日志**
   ```bash
   tail -f app.log | lg -u
   ```

5. **配合其他命令使用**
   ```bash
   cat *.log | lg -u | grep ERROR
   tail -n 100 app.log | lg -l
   ```


### 从源码构建

```bash
# 克隆项目
git clone <repository>
cd loglens

# 方法1: 使用 Makefile 构建（推荐）
make build          # 构建优化版本（文件大小约 2.1MB）
make install        # 安装到 /usr/local/bin/

# 方法2: 手动构建
go build -ldflags="-s -w" -o lg    # 优化版本
go build -o lg                      # 调试版本（文件大小约 3.1MB）

# 可选：安装到系统路径
sudo mv lg /usr/local/bin/
```

### 构建选项

**优化构建**（推荐）：
```bash
make build          # 使用 -ldflags="-s -w" 去除调试信息
                     # 文件大小: ~2.1MB（减小约 32%）
```

**调试构建**：
```bash
make build-debug    # 包含完整调试信息
                     # 文件大小: ~3.1MB
```

## 开发

### 使用 Makefile

```bash
# 查看所有可用命令
make help

# 构建
make build          # 优化构建
make build-debug    # 调试构建

# 清理
make clean

# 安装
make install        # 安装到系统
```


## 许可证

MIT License
