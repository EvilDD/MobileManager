# ➤ MobileManager

<div align="center">
<p align="center">
  <a href="https://www.bilibili.com/video/BV1bL5jz7E6y" target="_blank">
    观看合集视频-运行篇 | Watch Demo Video
  </a>
</p>

<p align="center">

[![License](https://img.shields.io/badge/License-CC%20BY--NC%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc/4.0/)
[![GoFrame](https://img.shields.io/badge/GoFrame-v2.0-brightgreen.svg)](https://goframe.org/)
[![Vue](https://img.shields.io/badge/Vue-3.x-green.svg)](https://vuejs.org/)
[![Pure Admin](https://img.shields.io/badge/Pure%20Admin-Latest-blue.svg)](https://github.com/pure-admin/vue-pure-admin)

</p>

<h4>基于 ADB 的移动设备管理系统 | ADB-based Mobile Device Management System</h4>

MobileManager 是一个强大的移动设备管理系统，基于 ADB (Android Debug Bridge) 实现设备控制和管理。系统采用 GoFrame + Vue3 Pure Admin 技术栈，提供设备管理、应用管理、远程控制等功能。

</div>

## 🚀 最近更新 (v3.0.2 - 2025-05-14)

- ✅ 云机同步功能增强
  - 添加手动横竖屏切换功能
  - 支持批量键盘输入和单个设备独立输入
  - 添加从设备放大显示操作
  - 支持对标点符号的键盘输入
- ✅ 云机同步性能优化
  - 优化批量启动和关闭逻辑，支持复用session
  - 优化端口转发并发，修复设备锁竞争问题
  - 优化日志显示，方便后期定位问题
- ✅ 串流功能优化
  - 修复分组列表页面串流自动横竖屏切换编码错误
  - 修复雷电与部分机型无法串流问题
  - 改进服务启动判断方法

查看完整更新历史：[更新日志](CHANGELOG.md)

## ✨ 核心特性

* 📱 设备管理
  - 添加、编辑、删除设备
  - 设备分组管理
  - 批量设备操作
  - 设备状态监控

* 📦 应用管理
  - 应用安装/卸载
  - 应用启动控制
  - 批量应用操作
  - 应用上传管理

* 🖥️ 远程控制
  - 设备实时串流
  - 设备远程操作
  - 低延迟传输

* 🎯 批量操作
  - 多设备并行控制
  - 任务状态实时显示
  - 操作结果反馈

## 🛠️ 技术栈

### 后端技术
- GoFrame v2.0：基于 Golang 的 Web 开发框架
- SQLite：轻量级数据库
- ADB：Android 调试桥接

### 前端技术
- Vue 3：渐进式 JavaScript 框架
- TypeScript：类型安全
- Pure Admin：优雅的后台管理模板
- Element Plus：UI 组件库
- Pinia：状态管理
- WebCodecs：视频解码技术

## 🚀 快速开始

### 环境要求
- Go 1.18+
- Node.js 16+
- ADB 工具（必须安装并配置环境变量）
- pnpm 包管理器

> ⚠️ **注意**：本系统依赖 ADB (Android Debug Bridge) 环境，请确保在使用前已正确安装并配置 ADB，且可以在命令行中直接使用 `adb` 命令。

### 安装步骤

1. 克隆项目
```bash
git clone https://github.com/yourusername/MobileManager.git
cd MobileManager
```

2. 后端服务
```bash
cd backend
go mod tidy
mv backend\manifest\config\config.yaml.bak backend\manifest\config\config.yaml
go run main.go
```

3. 前端服务
```bash
cd frontend
pnpm install
mv frontend\.env.bak frontend\.env.development
pnpm run dev
```

> 💡 **提示**：需要运行后端服务和前端服务。建议在两个不同的终端窗口中分别启动各服务。

## 📚 功能列表

### 已实现功能
- ✅ 设备管理（添加/编辑/删除）
- ✅ 设备分组管理
- ✅ 设备远程串流控制
- ✅ 批量设备操作
- ✅ 应用管理（上传/安装/卸载/启动）
- ✅ 图片缓存刷新
- ✅ 任务状态显示
- ✅ 云机同步（多设备同步操作）
- ✅ 基于WebCodecs的视频流解码
- ✅ 多设备主从画面显示
- ✅ 应用帐号管理
- ✅ 设备信息上报与管控
- ✅ 应用帐号编辑功能
- ✅ 2K以下分辨率屏幕显示支持
- ✅ 设备自动扫描一键添加

### 开发计划
- 🔲 云机同步横屏支持
- 🔲 脚本管理与执行
- 🔲 代理配置
- 🔲 设备性能监控
- 🔲 自动化测试支持
- 🔲 更多品牌设备兼容性支持
- 🔲 更高效的中大规模设备管理方案

## 📦 项目结构

```
MobileManager/
├── backend/                # GoFrame 后端项目
│   ├── api/               # API 接口定义
│   ├── internal/          # 内部实现
│   └── manifest/          # 配置文件
└── frontend/              # Vue3 前端项目
    ├── src/
    │   ├── api/          # API 请求
    │   ├── components/   # 组件
    │   └── views/        # 页面
```

## 📄 贡献指南

欢迎提交 Issue 和 Pull Request！

## 📄 开源协议

本项目采用 [Creative Commons Attribution-NonCommercial 4.0 International License (CC BY-NC 4.0)](https://creativecommons.org/licenses/by-nc/4.0/) 协议。

这意味着您可以：
- ✅ 自由使用、复制、修改和分享本项目
- ✅ 以任何形式重新分发本项目
- ❌ 不得将本项目用于商业目的

使用条件：
1. **署名**：必须给出适当的署名，提供指向本许可证的链接，同时标明是否对原始内容作出修改
2. **非商业性**：不得将本项目用于商业目的
3. **分享时保持许可协议一致**：如果您修改了本项目，必须以相同的许可证分发您的贡献
