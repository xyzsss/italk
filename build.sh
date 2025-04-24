#!/bin/bash

# 创建存放可执行文件的目录
mkdir -p dist

# 设置版本和构建时间
VERSION=$(git describe --tags --always || echo "unknown")
BUILD_TIME=$(date "+%F %T")
LDFLAGS="-s -w -X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'"

# 当前平台使用CGO（支持sqlite）
echo "打包当前平台版本（使用CGO支持SQLite）..."
export CGO_ENABLED=1
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)
if [ "$CURRENT_ARCH" = "x86_64" ]; then
    CURRENT_ARCH="amd64"
fi

if [ "$CURRENT_OS" = "darwin" ]; then
    go build -ldflags "$LDFLAGS" -o dist/chat-app-macos-$CURRENT_ARCH
    mkdir -p dist/chat-app-macos-$CURRENT_ARCH-package
    cp dist/chat-app-macos-$CURRENT_ARCH dist/chat-app-macos-$CURRENT_ARCH-package/
    cp -r static dist/chat-app-macos-$CURRENT_ARCH-package/
    cp -r templates dist/chat-app-macos-$CURRENT_ARCH-package/
    cd dist && tar -czf chat-app-macos-$CURRENT_ARCH-with-sqlite.tar.gz chat-app-macos-$CURRENT_ARCH-package && cd ..
    rm -rf dist/chat-app-macos-$CURRENT_ARCH-package
    echo "已创建本地MacOS版本（支持SQLite）: dist/chat-app-macos-$CURRENT_ARCH-with-sqlite.tar.gz"
elif [ "$CURRENT_OS" = "linux" ]; then
    go build -ldflags "$LDFLAGS" -o dist/chat-app-linux-$CURRENT_ARCH
    mkdir -p dist/chat-app-linux-$CURRENT_ARCH-package
    cp dist/chat-app-linux-$CURRENT_ARCH dist/chat-app-linux-$CURRENT_ARCH-package/
    cp -r static dist/chat-app-linux-$CURRENT_ARCH-package/
    cp -r templates dist/chat-app-linux-$CURRENT_ARCH-package/
    cd dist && tar -czf chat-app-linux-$CURRENT_ARCH-with-sqlite.tar.gz chat-app-linux-$CURRENT_ARCH-package && cd ..
    rm -rf dist/chat-app-linux-$CURRENT_ARCH-package
    echo "已创建本地Linux版本（支持SQLite）: dist/chat-app-linux-$CURRENT_ARCH-with-sqlite.tar.gz"
fi

# 跨平台版本，使用内存数据库
echo "修改为使用内存数据库以支持跨平台..."
export CGO_ENABLED=0

echo "打包Linux x64版本..."
export GOOS=linux
export GOARCH=amd64
go build -ldflags "$LDFLAGS" -o dist/chat-app-linux-amd64
mkdir -p dist/chat-app-linux-amd64-package
cp dist/chat-app-linux-amd64 dist/chat-app-linux-amd64-package/
cp -r static dist/chat-app-linux-amd64-package/
cp -r templates dist/chat-app-linux-amd64-package/
cd dist && tar -czf chat-app-linux-amd64.tar.gz chat-app-linux-amd64-package && cd ..
rm -rf dist/chat-app-linux-amd64-package

echo "打包MacOS Intel版本..."
export GOOS=darwin
export GOARCH=amd64
go build -ldflags "$LDFLAGS" -o dist/chat-app-macos-amd64
mkdir -p dist/chat-app-macos-amd64-package
cp dist/chat-app-macos-amd64 dist/chat-app-macos-amd64-package/
cp -r static dist/chat-app-macos-amd64-package/
cp -r templates dist/chat-app-macos-amd64-package/
cd dist && tar -czf chat-app-macos-amd64.tar.gz chat-app-macos-amd64-package && cd ..
rm -rf dist/chat-app-macos-amd64-package

echo "打包Windows x64版本..."
export GOOS=windows
export GOARCH=amd64
go build -ldflags "$LDFLAGS" -o dist/chat-app-windows-amd64.exe
mkdir -p dist/chat-app-windows-amd64-package
cp dist/chat-app-windows-amd64.exe dist/chat-app-windows-amd64-package/
cp -r static dist/chat-app-windows-amd64-package/
cp -r templates dist/chat-app-windows-amd64-package/
cd dist && zip -r chat-app-windows-amd64.zip chat-app-windows-amd64-package && cd ..
rm -rf dist/chat-app-windows-amd64-package

echo "打包完成！所有文件位于dist目录" 