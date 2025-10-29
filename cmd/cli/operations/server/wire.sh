#!/bin/zsh

# 检查 wire 是否已安装
if ! command -v wire &> /dev/null; then
    echo "Wire not found, installing..."
    go install github.com/google/wire/cmd/wire@latest
    if [ $? -ne 0 ]; then
        echo "Failed to install wire"
        exit 1
    fi
    echo "Wire installed successfully"
fi

# 运行 wire 生成代码
wire ./...
