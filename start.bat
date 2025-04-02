@echo off
chcp 65001
title 项目启动脚本

echo 正在启动所有服务...

:: 启动后端服务
echo 正在启动后端服务...
start "后端服务" cmd /c "cd backend && go run main.go"

:: 启动 wscrcpy 服务
echo 正在启动 wscrcpy 服务...
start "wscrcpy服务" cmd /c "cd wscrcpy && pnpm start"

:: 启动前端服务
echo 正在启动前端服务...
start "前端服务" cmd /c "cd frontend && pnpm run dev"

echo.
echo 所有服务已启动！
echo 请查看各个命令窗口的具体输出信息。
echo.
echo 按任意键退出此窗口...
pause > nul 