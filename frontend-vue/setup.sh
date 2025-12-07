#!/bin/bash

echo "🔧 NeverBroke 前端环境修复脚本"
echo "================================"
echo ""

# 检查 Node.js 版本
NODE_VERSION=$(node -v)
echo "📌 当前 Node.js 版本: $NODE_VERSION"
echo ""

# 修复 npm 缓存权限问题
echo "🔨 修复 npm 缓存权限..."
sudo chown -R $(whoami) "$HOME/.npm"
echo "✅ npm 缓存权限已修复"
echo ""

# 清理旧的依赖
echo "🧹 清理旧的依赖..."
rm -rf node_modules package-lock.json
echo "✅ 旧依赖已清理"
echo ""

# 清理 npm 缓存
echo "🧹 清理 npm 缓存..."
npm cache clean --force
echo "✅ npm 缓存已清理"
echo ""

# 安装依赖
echo "📦 安装依赖（兼容 Node.js 16）..."
npm install
echo ""

if [ $? -eq 0 ]; then
    echo "✅ 依赖安装成功！"
    echo ""
    echo "🚀 现在可以运行以下命令启动开发服务器："
    echo "   npm run dev"
    echo ""
    echo "💡 提示："
    echo "   - 如果仍有问题，建议升级到 Node.js 20 LTS"
    echo "   - 查看 NODE_VERSION.md 获取详细说明"
else
    echo "❌ 依赖安装失败"
    echo ""
    echo "📝 可能的解决方案："
    echo "   1. 升级 Node.js 到 20 LTS: nvm install 20 && nvm use 20"
    echo "   2. 查看 NODE_VERSION.md 获取详细说明"
    echo "   3. 手动执行: npm cache clean --force && npm install"
fi

