@echo off
setlocal

REM 创建存放可执行文件的目录
mkdir dist 2>nul

REM 确保使用原始环境变量
set CGO_ENABLED=0

REM 设置版本和构建时间
for /f "tokens=*" %%a in ('git describe --tags --always 2^>nul') do set VERSION=%%a
if "%VERSION%"=="" set VERSION=unknown
for /f "tokens=*" %%a in ('powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'"') do set BUILD_TIME=%%a
set LDFLAGS=-s -w -X "main.Version=%VERSION%" -X "main.BuildTime=%BUILD_TIME%"

echo 打包Linux x64版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o dist\chat-app-linux-amd64
mkdir dist\chat-app-linux-amd64-package 2>nul
copy dist\chat-app-linux-amd64 dist\chat-app-linux-amd64-package\
xcopy /E /I static dist\chat-app-linux-amd64-package\static
xcopy /E /I templates dist\chat-app-linux-amd64-package\templates
cd dist && tar -czf chat-app-linux-amd64.tar.gz chat-app-linux-amd64-package && cd ..
rmdir /S /Q dist\chat-app-linux-amd64-package

echo 打包MacOS Intel版本...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o dist\chat-app-macos-amd64
mkdir dist\chat-app-macos-amd64-package 2>nul
copy dist\chat-app-macos-amd64 dist\chat-app-macos-amd64-package\
xcopy /E /I static dist\chat-app-macos-amd64-package\static
xcopy /E /I templates dist\chat-app-macos-amd64-package\templates
cd dist && tar -czf chat-app-macos-amd64.tar.gz chat-app-macos-amd64-package && cd ..
rmdir /S /Q dist\chat-app-macos-amd64-package

echo 打包Windows x64版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o dist\chat-app-windows-amd64.exe
mkdir dist\chat-app-windows-amd64-package 2>nul
copy dist\chat-app-windows-amd64.exe dist\chat-app-windows-amd64-package\
xcopy /E /I static dist\chat-app-windows-amd64-package\static
xcopy /E /I templates dist\chat-app-windows-amd64-package\templates
cd dist && powershell -command "Compress-Archive -Path chat-app-windows-amd64-package -DestinationPath chat-app-windows-amd64.zip -Force" && cd ..
rmdir /S /Q dist\chat-app-windows-amd64-package

echo 打包完成！所有文件位于dist目录 