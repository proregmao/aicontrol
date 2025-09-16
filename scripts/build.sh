#!/bin/bash

# 智能设备管理系统构建脚本

set -e

echo "🏗️  构建智能设备管理系统..."

# 创建构建目录
mkdir -p dist

# 构建后端
build_backend() {
    echo "🔧 构建后端服务..."
    cd backend
    
    # 设置构建环境变量
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    # 构建二进制文件
    echo "📦 编译Go二进制文件..."
    go build -ldflags="-w -s" -o ../dist/smart-device-server cmd/server/main.go
    
    # 复制配置文件
    echo "📋 复制配置文件..."
    cp -r configs ../dist/
    
    cd ..
    echo "✅ 后端构建完成"
}

# 构建前端
build_frontend() {
    echo "🎨 构建前端应用..."
    cd frontend
    
    # 安装依赖（如果需要）
    if [ ! -d "node_modules" ]; then
        echo "📦 安装前端依赖..."
        npm install
    fi
    
    # 构建生产版本
    echo "🏗️  构建生产版本..."
    npm run build
    
    # 复制构建结果
    echo "📋 复制构建文件..."
    cp -r dist/* ../dist/
    
    cd ..
    echo "✅ 前端构建完成"
}

# 创建启动脚本
create_startup_script() {
    echo "📝 创建生产环境启动脚本..."
    
    cat > dist/start.sh << 'EOF'
#!/bin/bash

echo "🚀 启动智能设备管理系统..."

# 检查配置文件
if [ ! -f "configs/.env" ]; then
    echo "❌ 配置文件不存在，请复制 configs/.env.example 到 configs/.env 并配置"
    exit 1
fi

# 启动服务器
echo "🔧 启动服务器..."
./smart-device-server

EOF
    
    chmod +x dist/start.sh
    echo "✅ 启动脚本创建完成"
}

# 创建Docker文件
create_dockerfile() {
    echo "🐳 创建Dockerfile..."
    
    cat > dist/Dockerfile << 'EOF'
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制二进制文件和配置
COPY smart-device-server .
COPY configs ./configs
COPY . .

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./smart-device-server"]
EOF
    
    echo "✅ Dockerfile创建完成"
}

# 主执行流程
main() {
    echo "🔄 开始构建流程..."
    
    build_backend
    build_frontend
    create_startup_script
    create_dockerfile
    
    echo ""
    echo "🎉 构建完成！"
    echo ""
    echo "📁 构建文件位置: ./dist/"
    echo "🚀 启动命令: cd dist && ./start.sh"
    echo "🐳 Docker构建: cd dist && docker build -t smart-device-management ."
    echo ""
}

# 执行主函数
main
