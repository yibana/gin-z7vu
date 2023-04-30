#!/bin/bash

# 更新代码
echo "Updating repository..."
git pull

# 编译
echo "Building binary..."
go build -o gin

# 启动
echo "Starting application..."
./gin
