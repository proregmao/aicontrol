#!/bin/bash

# 智能设备管理系统数据库初始化脚本

set -e

echo "🗄️  初始化智能设备管理系统数据库..."

# 默认配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-smart_device_management}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}

# 检查PostgreSQL是否可用
check_postgres() {
    echo "🔍 检查PostgreSQL连接..."
    
    if ! command -v psql &> /dev/null; then
        echo "❌ PostgreSQL客户端未安装"
        echo "请安装PostgreSQL客户端: sudo apt-get install postgresql-client"
        exit 1
    fi
    
    # 测试连接
    if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" &> /dev/null; then
        echo "❌ 无法连接到PostgreSQL服务器"
        echo "请检查数据库服务器是否运行，以及连接参数是否正确"
        echo "连接参数: $DB_USER@$DB_HOST:$DB_PORT"
        exit 1
    fi
    
    echo "✅ PostgreSQL连接正常"
}

# 创建数据库
create_database() {
    echo "📝 创建数据库 $DB_NAME..."
    
    # 检查数据库是否已存在
    if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
        echo "ℹ️  数据库 $DB_NAME 已存在"
        
        read -p "是否要重新创建数据库？这将删除所有现有数据 (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "🗑️  删除现有数据库..."
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
            echo "📝 创建新数据库..."
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        fi
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        echo "✅ 数据库创建成功"
    fi
}

# 初始化数据库结构
init_database() {
    echo "🏗️  初始化数据库结构..."
    
    # 执行初始化SQL脚本
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f scripts/init-db.sql
    
    echo "✅ 数据库结构初始化完成"
}

# 验证数据库
verify_database() {
    echo "🔍 验证数据库初始化..."
    
    # 检查表是否创建成功
    table_count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
    
    if [ "$table_count" -gt 0 ]; then
        echo "✅ 数据库验证成功，共创建 $table_count 个表"
        
        # 显示表列表
        echo "📋 数据库表列表:"
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\dt"
    else
        echo "❌ 数据库验证失败，未找到任何表"
        exit 1
    fi
}

# 创建环境变量文件
create_env_file() {
    echo "📝 创建数据库环境变量文件..."
    
    ENV_FILE="backend/configs/.env"
    
    if [ ! -f "$ENV_FILE" ]; then
        cp backend/configs/.env.example "$ENV_FILE"
    fi
    
    # 更新数据库配置
    sed -i "s/DB_HOST=.*/DB_HOST=$DB_HOST/" "$ENV_FILE"
    sed -i "s/DB_PORT=.*/DB_PORT=$DB_PORT/" "$ENV_FILE"
    sed -i "s/DB_NAME=.*/DB_NAME=$DB_NAME/" "$ENV_FILE"
    sed -i "s/DB_USER=.*/DB_USER=$DB_USER/" "$ENV_FILE"
    sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=$DB_PASSWORD/" "$ENV_FILE"
    
    echo "✅ 环境变量文件已更新: $ENV_FILE"
}

# 显示使用说明
show_usage() {
    echo ""
    echo "🎉 数据库初始化完成！"
    echo ""
    echo "📊 数据库信息:"
    echo "   主机: $DB_HOST:$DB_PORT"
    echo "   数据库: $DB_NAME"
    echo "   用户: $DB_USER"
    echo ""
    echo "👤 默认管理员账户:"
    echo "   用户名: admin"
    echo "   密码: admin123"
    echo ""
    echo "🚀 下一步:"
    echo "   1. 启动后端服务: cd backend && go run cmd/server/main.go"
    echo "   2. 启动前端服务: cd frontend && npm run dev"
    echo "   3. 或使用开发脚本: ./scripts/dev-start.sh"
    echo ""
}

# 主执行流程
main() {
    echo "🔧 数据库配置:"
    echo "   主机: $DB_HOST:$DB_PORT"
    echo "   数据库: $DB_NAME"
    echo "   用户: $DB_USER"
    echo ""
    
    check_postgres
    create_database
    init_database
    verify_database
    create_env_file
    show_usage
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --host HOST         数据库主机 (默认: localhost)"
            echo "  --port PORT         数据库端口 (默认: 5432)"
            echo "  --database NAME     数据库名称 (默认: smart_device_management)"
            echo "  --user USER         数据库用户 (默认: postgres)"
            echo "  --password PASS     数据库密码 (默认: password)"
            echo "  --help              显示此帮助信息"
            echo ""
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 执行主函数
main
