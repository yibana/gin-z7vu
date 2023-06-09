#!/bin/bash

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

# 更新代码
echo "Updating repository..."
git reset --hard HEAD
git pull
# 编译
echo "Building binary..."
go build -o gin main.go



# Set environment variable
export MONGO_URL="mongodb://admin:3WPIki9dXShd6ZZhGXKZ@127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+1.8.2"
export REDIS_URL="redis://localhost:6379"
chmod +x pull_deploy.sh

# Start gin and log output to gin.log
nohup ./gin > gin.log 2>&1 &
# View logs
tail -f gin.log
