# Git AI 工具

一个基于 Wails 的 Git 桌面管理工具，支持 AI 辅助生成 commit 信息。

## 功能特性

### 仓库管理
- **打开仓库** - 打开本地 Git 仓库
- **克隆仓库** - 克隆远程仓库到本地
- **仓库列表** - 管理常用仓库，快速切换
- **最近打开** - 快速访问最近使用的仓库

### Git 操作
- **状态查看** - 查看工作区、暂存区变更
- **分支管理** - 创建、切换、删除分支
- **提交历史** - 查看项目提交记录
- **标签管理** - 创建、删除标签

### 远程操作
- **推送** - 推送代码到远程仓库
- **拉取** - 从远程仓库拉取更新

### 版本控制
- **撤销提交** - 软撤销、混合撤销、硬撤销
- **回滚** - 回滚到指定提交

### AI 辅助
- **AI 配置** - 配置 AI 服务（OpenAI 兼容 API）
- **智能提示词** - 自定义 commit 生成提示词
- **自动生成** - 根据变更自动生成 commit 信息

## 快速开始

### 环境要求
- Go 1.19+
- Node.js 16+
- Wails CLI

### 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install
```

### 开发模式

```bash
# 在项目根目录运行
wails dev
```

### 构建发布版本

```bash
wails build
```

## 配置说明

### AI 配置
在设置中配置 AI 服务：
- **Provider** - AI 服务提供商
- **Base URL** - API 地址
- **API Key** - API 密钥
- **Model** - 模型名称

### 提示词模板
可自定义 commit 信息生成规则，支持以下变量：
- `{{.diff}}` - 代码变更内容
- `{{.branch}}` - 分支名称
- `{{.message}}` - 最近提交信息

## 技术栈

- **前端**: Vue 3 + TypeScript + Vite
- **后端**: Go + Wails
- **UI**: 自定义深色主题

## 目录结构

```
git-ai-tools/
├── app.go              # 应用入口
├── main.go             # 主程序
├── internal/
│   ├── ai/            # AI 服务
│   ├── config/        # 配置管理
│   ├── git/           # Git 操作
│   └── models/        # 数据模型
├── frontend/          # Vue 前端
│   ├── src/
│   │   ├── components/ # 组件
│   │   └── App.vue     # 主应用
│   └── package.json
└── wails.json         # Wails 配置
```

## License

MIT
