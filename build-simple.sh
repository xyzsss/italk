#!/bin/bash

# 设置版本和构建时间
VERSION="dev"
BUILD_TIME=$(date "+%F %T")
LDFLAGS="-s -w -X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'"

# 创建目录
mkdir -p dist

echo "使用内存数据库模式构建（不依赖CGO）..."
export CGO_ENABLED=0

echo "为当前平台构建中..."
go build -ldflags "$LDFLAGS" -o dist/chat-app

echo "复制静态资源..."
mkdir -p dist/package
cp dist/chat-app dist/package/
cp -r static dist/package/
cp -r templates dist/package/

echo "打包完成！可执行文件位于 dist/chat-app" 