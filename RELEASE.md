# 发布指南

本文档说明如何发布新版本的 LogLens。

## 发布流程

### 1. 更新版本号

确保代码已提交并推送到主分支。

### 2. 创建并推送 tag

```bash
# 创建 tag（例如 v1.0.0）
git tag -a v1.0.0 -m "Release version 1.0.0"

# 推送 tag 到远程仓库
git push origin v1.0.0
```

### 3. 自动构建和发布

推送 tag 后，GitHub Actions 会自动执行以下操作：

1. **构建多平台二进制文件**：
   - Linux AMD64
   - Linux ARM64
   - macOS AMD64 (Intel)
   - macOS ARM64 (Apple Silicon)
   - Windows AMD64
   - Windows ARM64

2. **创建压缩包**：
   - Linux/macOS: `.tar.gz` 格式
   - Windows: `.zip` 格式

3. **生成校验和**：
   - 自动生成 `checksums.txt` 文件

4. **创建 GitHub Release**：
   - 自动创建发布页面
   - 上传所有构建的二进制文件
   - 生成发布说明

### 4. 验证发布

访问 [GitHub Releases](https://github.com/YOUR_USERNAME/loglens/releases) 页面，确认：

- ✅ Release 已创建
- ✅ 所有平台的二进制文件已上传
- ✅ checksums.txt 已生成
- ✅ Release notes 正确生成

## 版本号规范

遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- **主版本号（MAJOR）**：不兼容的 API 修改
- **次版本号（MINOR）**：向下兼容的功能性新增
- **修订号（PATCH）**：向下兼容的问题修正

示例：
- `v1.0.0` - 首个稳定版本
- `v1.1.0` - 添加新功能
- `v1.1.1` - 修复 bug
- `v2.0.0` - 重大更新，可能不兼容旧版本

## 构建的平台和架构

| 平台 | 架构 | 文件名 |
|------|------|--------|
| Linux | AMD64 | `lg-VERSION-linux-amd64.tar.gz` |
| Linux | ARM64 | `lg-VERSION-linux-arm64.tar.gz` |
| macOS | AMD64 (Intel) | `lg-VERSION-darwin-amd64.tar.gz` |
| macOS | ARM64 (M1/M2) | `lg-VERSION-darwin-arm64.tar.gz` |
| Windows | AMD64 | `lg-VERSION-windows-amd64.zip` |
| Windows | ARM64 | `lg-VERSION-windows-arm64.zip` |

## 发布前检查清单

- [ ] 所有测试通过
- [ ] 代码已格式化（`make fmt`）
- [ ] 代码检查通过（`make vet`）
- [ ] README 已更新
- [ ] CHANGELOG 已更新（如果有）
- [ ] 版本号符合语义化版本规范
- [ ] 本地测试构建成功

## 发布后任务

- [ ] 验证所有平台的二进制文件可以正常运行
- [ ] 更新文档中的版本号引用
- [ ] 在社交媒体或论坛宣布新版本（可选）
- [ ] 关闭相关的 issue 和 milestone（如果有）

## 紧急回滚

如果发现严重问题需要回滚：

```bash
# 删除远程 tag
git push --delete origin v1.0.0

# 删除本地 tag
git tag -d v1.0.0

# 在 GitHub 上手动删除 Release
```

然后修复问题，重新发布新的版本。

## 示例

完整的发布流程示例：

```bash
# 1. 确保代码最新
git checkout main
git pull origin main

# 2. 运行测试
make test
make vet
make fmt

# 3. 本地构建测试
make build

# 4. 创建 tag
git tag -a v1.0.0 -m "Release version 1.0.0

Features:
- 交互式日志查看
- JSON 格式化支持
- 搜索和高亮功能
- 多平台支持"

# 5. 推送 tag
git push origin v1.0.0

# 6. 等待 GitHub Actions 完成构建（约 3-5 分钟）

# 7. 访问 Releases 页面验证
# https://github.com/YOUR_USERNAME/loglens/releases
```

## 故障排除

### 构建失败

查看 [GitHub Actions](https://github.com/YOUR_USERNAME/loglens/actions) 页面的详细日志。

常见问题：
- Go 版本不匹配
- 依赖项缺失
- 权限问题

### Release 未创建

确保：
- GitHub Token 权限正确
- Workflow 文件语法正确
- Tag 格式符合 `v*` 模式

## 参考链接

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Go 跨平台编译](https://golang.org/doc/install/source#environment)
- [语义化版本规范](https://semver.org/lang/zh-CN/)
