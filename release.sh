#!/bin/bash

# LogLens 发布脚本
# 用法: ./release.sh <version>
# 示例: ./release.sh v1.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查参数
if [ $# -eq 0 ]; then
    echo -e "${RED}错误: 请提供版本号${NC}"
    echo "用法: $0 <version>"
    echo "示例: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# 验证版本号格式
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}错误: 版本号格式不正确${NC}"
    echo "版本号应该符合 vX.Y.Z 格式，例如 v1.0.0"
    exit 1
fi

echo -e "${GREEN}=== LogLens 发布脚本 ===${NC}"
echo -e "版本: ${YELLOW}${VERSION}${NC}"
echo ""

# 1. 检查工作目录状态
echo -e "${GREEN}[1/7] 检查工作目录状态...${NC}"
if [[ -n $(git status -s) ]]; then
    echo -e "${RED}错误: 工作目录有未提交的更改${NC}"
    git status -s
    exit 1
fi
echo -e "${GREEN}✓ 工作目录干净${NC}"
echo ""

# 2. 确保在主分支
echo -e "${GREEN}[2/7] 检查分支...${NC}"
BRANCH=$(git branch --show-current)
if [[ "$BRANCH" != "main" && "$BRANCH" != "master" ]]; then
    echo -e "${YELLOW}警告: 当前不在 main/master 分支 (当前: $BRANCH)${NC}"
    read -p "是否继续? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
echo -e "${GREEN}✓ 分支: $BRANCH${NC}"
echo ""

# 3. 拉取最新代码
echo -e "${GREEN}[3/7] 拉取最新代码...${NC}"
git pull origin $BRANCH
echo -e "${GREEN}✓ 代码已更新${NC}"
echo ""

# 4. 运行测试
echo -e "${GREEN}[4/7] 运行测试...${NC}"
if command -v make &> /dev/null; then
    make test || { echo -e "${RED}✗ 测试失败${NC}"; exit 1; }
    make vet || { echo -e "${RED}✗ 代码检查失败${NC}"; exit 1; }
else
    go test ./... || { echo -e "${RED}✗ 测试失败${NC}"; exit 1; }
    go vet ./... || { echo -e "${RED}✗ 代码检查失败${NC}"; exit 1; }
fi
echo -e "${GREEN}✓ 所有测试通过${NC}"
echo ""

# 5. 检查 tag 是否已存在
echo -e "${GREEN}[5/7] 检查标签...${NC}"
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}错误: 标签 $VERSION 已存在${NC}"
    echo "如需重新发布，请先删除现有标签:"
    echo "  git tag -d $VERSION"
    echo "  git push --delete origin $VERSION"
    exit 1
fi
echo -e "${GREEN}✓ 标签可用${NC}"
echo ""

# 6. 创建标签
echo -e "${GREEN}[6/7] 创建标签 $VERSION...${NC}"
read -p "请输入发布说明 (可选，直接回车跳过): " RELEASE_NOTES
if [[ -z "$RELEASE_NOTES" ]]; then
    git tag -a "$VERSION" -m "Release $VERSION"
else
    git tag -a "$VERSION" -m "$RELEASE_NOTES"
fi
echo -e "${GREEN}✓ 标签已创建${NC}"
echo ""

# 7. 推送标签
echo -e "${GREEN}[7/7] 推送标签到远程仓库...${NC}"
echo -e "${YELLOW}这将触发 GitHub Actions 自动构建和发布${NC}"
read -p "确认推送? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}已取消。标签已在本地创建，可以稍后手动推送:${NC}"
    echo "  git push origin $VERSION"
    exit 0
fi

git push origin "$VERSION"
echo -e "${GREEN}✓ 标签已推送${NC}"
echo ""

# 完成
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}发布流程已启动！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "接下来的步骤:"
echo "1. GitHub Actions 正在构建多平台二进制文件"
echo "2. 访问 Actions 页面查看构建进度:"
echo -e "   ${YELLOW}https://github.com/YOUR_USERNAME/loglens/actions${NC}"
echo "3. 构建完成后，Release 将自动发布:"
echo -e "   ${YELLOW}https://github.com/YOUR_USERNAME/loglens/releases/tag/$VERSION${NC}"
echo ""
echo "预计构建时间: 3-5 分钟"
echo ""
echo -e "${GREEN}✓ 完成！${NC}"
