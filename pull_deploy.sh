#!/bin/bash

# 更新代码
echo "Updating repository..."
git reset --hard HEAD
git pull

# 编译
echo "Building binary..."
go build -o gin

# Check if the "-kill" parameter is passed
if [ "$1" = "-kill" ]; then
  # Kill the existing gin process if it's running
  if pgrep gin >/dev/null 2>&1; then
    echo "Stopping existing gin process..."
    pkill -f gin
  else
    echo "gin process is not running"
  fi
  exit 0
fi

# Check if gin process is already running, and end the process
if pgrep gin >/dev/null 2>&1; then
  echo "Stopping existing gin process..."
  pkill -f gin
fi

# Set environment variable
export MONGO_URL="mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+1.8.2"

# Start gin and log output to gin.log
nohup ./gin > gin.log 2>&1 &

# View logs
tail -f gin.log
