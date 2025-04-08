package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "backend/api/scrcpy/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

type sScrcpy struct {
	sessions      map[string]*StreamSession
	sessionsLock  sync.RWMutex
	serverJarPath string
	startTimeout  time.Duration // 启动超时时间
	maxRetries    int           // 最大重试次数
	retryInterval time.Duration // 重试间隔
	debugLog      bool          // 是否输出详细日志
}

type StreamSession struct {
	DeviceId  string
	Port      int
	ProcessId int
	LocalPort int // 本地转发端口
}

func init() {
	localScrcpy = New()
}

func New() *sScrcpy {
	return &sScrcpy{
		sessions:      make(map[string]*StreamSession),
		serverJarPath: "resource/scrcpy/scrcpy-server.jar",
		startTimeout:  30 * time.Second,       // 默认30秒超时
		maxRetries:    5,                      // 默认最多重试5次
		retryInterval: 500 * time.Millisecond, // 默认500ms重试间隔
		debugLog:      true,                   // 默认开启调试日志
	}
}

// StartStream 启动设备流服务
func (s *sScrcpy) StartStream(ctx context.Context, req *v1.StartStreamReq) (res *v1.StartStreamRes, err error) {
	// 先检查设备状态
	if err := s.checkDeviceStatus(ctx, req.DeviceId); err != nil {
		return nil, err
	}

	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	// 检查会话是否已存在
	if session, exists := s.sessions[req.DeviceId]; exists {
		s.logDebug("发现已存在的会话，设备: %s, 端口: %d", req.DeviceId, session.LocalPort)

		// 验证会话是否仍然有效
		if s.validateExistingSession(ctx, session) {
			s.logDebug("现有会话有效，直接返回")
			return &v1.StartStreamRes{
				Port: session.LocalPort,
				Url:  fmt.Sprintf("ws://localhost:%d", session.LocalPort),
			}, nil
		}

		// 会话无效，需要清理旧会话并创建新会话
		s.logDebug("现有会话无效，将创建新会话")
		s.killServer(ctx, session)
		s.removePortForward(ctx, session)
		delete(s.sessions, req.DeviceId)
	}

	// 创建新会话
	session := &StreamSession{
		DeviceId: req.DeviceId,
		Port:     8886, // 设备端端口，保持固定
	}

	// 推送服务器文件
	if err := s.pushServer(ctx, req.DeviceId); err != nil {
		return nil, err
	}

	// 启动服务器
	if err := s.startServer(ctx, session); err != nil {
		// 检查是否是端口占用错误
		if strings.Contains(err.Error(), "Address already in use") {
			return nil, gerror.New("设备正在被其他用户使用中")
		}
		return nil, err
	}

	// 设置ADB端口转发
	localPort, err := s.setupPortForward(ctx, session)
	if err != nil {
		s.killServer(ctx, session)
		return nil, gerror.Wrap(err, "Failed to setup port forwarding")
	}

	session.LocalPort = localPort
	s.sessions[req.DeviceId] = session

	// 最终验证端口是否可达
	if !s.isLocalPortReachable(localPort) {
		s.logDebug("本地端口无法访问，将清理会话")
		s.killServer(ctx, session)
		s.removePortForward(ctx, session)
		delete(s.sessions, req.DeviceId)
		return nil, gerror.New("服务启动成功，但本地端口无法访问")
	}

	s.logDebug("流服务启动成功: 设备=%s, 本地端口=%d", req.DeviceId, localPort)
	return &v1.StartStreamRes{
		Port: localPort,
		Url:  fmt.Sprintf("ws://localhost:%d", localPort),
	}, nil
}

// StopStream 停止设备流服务
func (s *sScrcpy) StopStream(ctx context.Context, req *v1.StopStreamReq) (res *v1.StopStreamRes, err error) {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	session, exists := s.sessions[req.DeviceId]
	if !exists {
		return &v1.StopStreamRes{}, nil
	}

	// 移除端口转发
	if session.LocalPort > 0 {
		s.removePortForward(ctx, session)
	}

	// 停止服务器进程
	if err := s.killServer(ctx, session); err != nil {
		return nil, err
	}

	delete(s.sessions, req.DeviceId)
	return &v1.StopStreamRes{}, nil
}

// 改进 pushServer 方法
func (s *sScrcpy) pushServer(ctx context.Context, deviceId string) error {
	jarPath := gfile.Join(gfile.MainPkgPath(), s.serverJarPath)
	if !gfile.Exists(jarPath) {
		return gerror.Newf("scrcpy-server.jar not found at: %s", jarPath)
	}

	s.logDebug("推送服务器文件: %s -> 设备: %s", jarPath, deviceId)

	// 1. 先检查目标设备上是否已存在服务器文件，并验证其大小
	checkCmd := exec.Command("adb", "-s", deviceId, "shell",
		"ls -l /data/local/tmp/scrcpy-server.jar 2>/dev/null || echo 'NOT_FOUND'")

	checkOutput, err := checkCmd.Output()
	checkResult := strings.TrimSpace(string(checkOutput))

	// 获取本地文件大小
	localFileSize := gfile.Size(jarPath)
	s.logDebug("本地JAR文件大小: %d 字节", localFileSize)

	// 判断是否需要重新推送
	needPush := true
	if err == nil && !strings.Contains(checkResult, "NOT_FOUND") {
		// 文件存在，检查大小
		s.logDebug("设备上已存在JAR文件: %s", checkResult)

		// 从ls -l输出解析文件大小
		parts := strings.Fields(checkResult)
		if len(parts) >= 5 {
			remoteSize := 0
			_, err := fmt.Sscanf(parts[4], "%d", &remoteSize)
			if err == nil && remoteSize > 0 {
				s.logDebug("设备JAR文件大小: %d 字节", remoteSize)

				// 如果大小相同，可以跳过推送(允许1KB的误差)
				localSizeInt := int(localFileSize) // 转换为int以便比较
				if remoteSize >= localSizeInt-1024 && remoteSize <= localSizeInt+1024 {
					s.logDebug("文件大小匹配，跳过推送")
					needPush = false
				}
			}
		}
	} else {
		s.logDebug("设备上不存在JAR文件，需要推送")
	}

	if !needPush {
		// 验证文件权限
		permCmd := exec.Command("adb", "-s", deviceId, "shell",
			"[ -r /data/local/tmp/scrcpy-server.jar ] && echo 'OK' || echo 'NO_PERMISSION'")
		permOutput, _ := permCmd.Output()

		if strings.TrimSpace(string(permOutput)) != "OK" {
			s.logDebug("文件存在但权限不正确，设置权限")
			// 设置权限
			chmodCmd := exec.Command("adb", "-s", deviceId, "shell",
				"chmod 644 /data/local/tmp/scrcpy-server.jar")
			_ = chmodCmd.Run()
		} else {
			s.logDebug("文件已存在且权限正确")
			return nil
		}
	}

	// 2. 执行推送操作
	s.logDebug("开始推送JAR文件...")
	cmd := exec.Command("adb", "-s", deviceId, "push", jarPath, "/data/local/tmp/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		s.logDebug("推送失败: %v，尝试恢复...", err)
		// 推送失败时尝试重试
		return s.pushServerWithRetry(ctx, deviceId, jarPath)
	}

	// 3. 验证推送结果
	verifyCmd := exec.Command("adb", "-s", deviceId, "shell",
		"[ -f /data/local/tmp/scrcpy-server.jar ] && echo 'EXISTS' || echo 'MISSING'")
	verifyOutput, _ := verifyCmd.Output()

	if strings.TrimSpace(string(verifyOutput)) != "EXISTS" {
		s.logDebug("推送后验证失败，文件不存在")
		return gerror.New("推送文件后验证失败，文件不存在")
	}

	s.logDebug("JAR文件推送成功")
	return nil
}

// 添加带重试机制的推送方法
func (s *sScrcpy) pushServerWithRetry(ctx context.Context, deviceId string, jarPath string) error {
	// 最多尝试3次
	for i := 0; i < 3; i++ {
		s.logDebug("重试推送 #%d...", i+1)

		// 先尝试清理之前可能存在的损坏文件
		cleanCmd := exec.Command("adb", "-s", deviceId, "shell",
			"rm -f /data/local/tmp/scrcpy-server.jar")
		_ = cleanCmd.Run()

		// 尝试使用不同参数推送
		cmd := exec.Command("adb", "-s", deviceId, "push", "--sync", jarPath, "/data/local/tmp/")

		if err := cmd.Run(); err != nil {
			s.logDebug("重试 #%d 失败: %v", i+1, err)
			// 等待一段时间再重试
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
			continue
		}

		// 验证推送结果
		verifyCmd := exec.Command("adb", "-s", deviceId, "shell",
			"[ -f /data/local/tmp/scrcpy-server.jar ] && echo 'EXISTS' || echo 'MISSING'")
		verifyOutput, _ := verifyCmd.Output()

		if strings.TrimSpace(string(verifyOutput)) == "EXISTS" {
			// 设置适当的权限
			chmodCmd := exec.Command("adb", "-s", deviceId, "shell",
				"chmod 644 /data/local/tmp/scrcpy-server.jar")
			_ = chmodCmd.Run()

			s.logDebug("重试推送成功")
			return nil
		}
	}

	return gerror.New("推送文件失败，多次重试后仍无法完成")
}

func (s *sScrcpy) startServer(ctx context.Context, session *StreamSession) error {
	// 1. 检查设备上是否已经有scrcpy服务器进程在运行
	pid, err := s.getServerPid(ctx, session.DeviceId)
	if err != nil {
		return gerror.Wrap(err, "Failed to check server process")
	}

	// 如果已有服务器进程在运行，复用该进程
	if pid > 0 {
		session.ProcessId = pid
		return nil
	}

	// 2. 启动服务器命令
	cmd := exec.Command("adb", "-s", session.DeviceId, "shell",
		fmt.Sprintf("CLASSPATH=/data/local/tmp/scrcpy-server.jar nohup app_process / com.genymobile.scrcpy.Server 1.19-ws6 web ERROR %d true 2>&1 > /dev/null",
			session.Port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return gerror.Wrap(err, "Failed to start scrcpy server")
	}

	// 3. 在后台监控进程
	go func() {
		cmd.Wait()
		s.sessionsLock.Lock()
		delete(s.sessions, session.DeviceId)
		s.sessionsLock.Unlock()
	}()

	// 4. 等待服务器启动并验证PID
	startErr := s.waitForServerStart(ctx, session)
	if startErr != nil {
		// 如果启动失败，尝试清理
		s.killServer(ctx, session)
		return gerror.Wrap(startErr, "Server failed to start properly")
	}

	return nil
}

// 获取服务器进程ID
func (s *sScrcpy) getServerPid(ctx context.Context, deviceId string) (int, error) {
	// 使用ps命令查找服务器进程
	cmd := exec.Command("adb", "-s", deviceId, "shell",
		"ps -ef | grep 'app_process.*com.genymobile.scrcpy.Server.*web.*ERROR' | grep -v grep | awk '{print $2}'")

	output, err := cmd.Output()
	if err != nil {
		// 忽略退出状态为1的错误，通常表示没有找到匹配进程
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return 0, nil
		}
		return 0, gerror.Wrap(err, "Failed to check server process")
	}

	// 解析进程ID
	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0, nil
	}

	// 可能有多行输出，取第一个有效的PID
	pids := strings.Split(pidStr, "\n")
	for _, pid := range pids {
		pid = strings.TrimSpace(pid)
		if pid != "" {
			pidInt := 0
			_, err := fmt.Sscanf(pid, "%d", &pidInt)
			if err == nil && pidInt > 0 {
				return pidInt, nil
			}
		}
	}

	return 0, nil
}

// 添加调试日志函数
func (s *sScrcpy) logDebug(format string, args ...interface{}) {
	if s.debugLog {
		fmt.Printf("[SCRCPY DEBUG] "+format+"\n", args...)
	}
}

// 改进等待服务器启动函数，增加更可靠的检测逻辑
func (s *sScrcpy) waitForServerStart(ctx context.Context, session *StreamSession) error {
	// 使用配置的timeout和重试参数
	maxTries := s.maxRetries
	s.logDebug("等待服务器启动，设备: %s, 端口: %d, 最大尝试次数: %d", session.DeviceId, session.Port, maxTries)

	// 创建带超时的子上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, s.startTimeout)
	defer cancel()

	for i := 0; i < maxTries; i++ {
		select {
		case <-timeoutCtx.Done():
			return gerror.New("服务器启动超时")
		default:
			// 继续执行
		}

		// 1. 检查服务器进程
		pid, err := s.getServerPid(ctx, session.DeviceId)
		if err != nil {
			s.logDebug("获取进程ID失败: %v，重试中...", err)
		} else if pid > 0 {
			s.logDebug("找到服务器进程 PID: %d", pid)
			// 找到有效进程ID
			session.ProcessId = pid

			// 2. 使用多种方法检查端口状态
			portOpen := false

			// 方法1: 使用netstat命令
			if s.isPortOpen(ctx, session.DeviceId, session.Port) {
				s.logDebug("端口检测成功(netstat): %d", session.Port)
				portOpen = true
			} else {
				// 方法2: 尝试检查进程打开的文件描述符
				if s.checkProcessSocket(ctx, session.DeviceId, pid, session.Port) {
					s.logDebug("端口检测成功(lsof): %d", session.Port)
					portOpen = true
				} else {
					// 方法3: 尝试直接连接测试
					if s.testConnection(ctx, session.DeviceId, session.Port) {
						s.logDebug("端口检测成功(连接测试): %d", session.Port)
						portOpen = true
					}
				}
			}

			if portOpen {
				s.logDebug("服务器启动成功！PID: %d，端口: %d", pid, session.Port)
				return nil // 成功启动
			}

			s.logDebug("服务器进程存在但端口未开放，等待中...")
		} else {
			s.logDebug("未找到服务器进程，尝试 %d/%d", i+1, maxTries)
		}

		// 等待间隔时间再检查
		sleepTime := s.retryInterval + time.Duration(i*300)*time.Millisecond // 递增等待时间
		time.Sleep(sleepTime)
	}

	// 如果所有尝试都失败，尝试重启服务一次
	s.logDebug("常规启动超时，尝试强制杀死现有进程并重启服务...")
	s.killServer(ctx, session)

	// 重新启动服务器
	cmd := exec.Command("adb", "-s", session.DeviceId, "shell",
		fmt.Sprintf("CLASSPATH=/data/local/tmp/scrcpy-server.jar nohup app_process / com.genymobile.scrcpy.Server 1.19-ws6 web ERROR %d true 2>&1 > /dev/null",
			session.Port))

	if err := cmd.Start(); err != nil {
		return gerror.Wrap(err, "重试启动服务器失败")
	}

	// 等待额外的时间
	time.Sleep(3 * time.Second)

	// 最后检查
	pid, _ := s.getServerPid(ctx, session.DeviceId)
	if pid > 0 {
		session.ProcessId = pid
		s.logDebug("重试成功！服务器已启动，PID: %d", pid)
		return nil
	}

	return gerror.New("服务器启动超时，请检查设备连接和ADB状态")
}

// 添加额外的端口检测方法：使用lsof检查进程打开的套接字
func (s *sScrcpy) checkProcessSocket(ctx context.Context, deviceId string, pid int, port int) bool {
	// 使用lsof命令检查进程是否打开了指定端口
	cmd := exec.Command("adb", "-s", deviceId, "shell",
		fmt.Sprintf("ls -l /proc/%d/fd/ 2>/dev/null | grep socket", pid))

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// 如果有socket文件描述符，可能表示服务已启动
	return len(output) > 0
}

// 添加直接连接测试方法
func (s *sScrcpy) testConnection(ctx context.Context, deviceId string, port int) bool {
	// 创建一个临时TCP连接测试端口是否可用
	cmd := exec.Command("adb", "-s", deviceId, "shell",
		fmt.Sprintf("(echo >/dev/tcp/localhost/%d) 2>/dev/null && echo 'open' || echo 'closed'", port))

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "open"
}

// 改进isPortOpen方法，使其更可靠
func (s *sScrcpy) isPortOpen(ctx context.Context, deviceId string, port int) bool {
	// 尝试多种端口检测命令，增加可靠性
	cmds := []struct {
		name string
		args []string
	}{
		{
			name: "netstat",
			args: []string{"-s", deviceId, "shell", fmt.Sprintf("netstat -tlnp 2>/dev/null | grep ':%d'", port)},
		},
		{
			name: "ss",
			args: []string{"-s", deviceId, "shell", fmt.Sprintf("ss -tlnp 2>/dev/null | grep ':%d'", port)},
		},
		{
			name: "lsof",
			args: []string{"-s", deviceId, "shell", fmt.Sprintf("lsof -i :%d 2>/dev/null", port)},
		},
	}

	for _, c := range cmds {
		cmd := exec.Command("adb", c.args...)
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			s.logDebug("端口 %d 检测成功（使用%s）", port, c.name)
			return true
		}
	}

	return false
}

func (s *sScrcpy) killServer(ctx context.Context, session *StreamSession) error {
	// 首先检查进程是否真的需要被终止
	if session.ProcessId > 0 {
		// 检查进程是否仍然在运行且是有效的scrcpy服务器进程
		cmd := exec.Command("adb", "-s", session.DeviceId, "shell",
			fmt.Sprintf("ps -p %d -o comm=", session.ProcessId))
		output, err := cmd.Output()
		if err == nil {
			processName := strings.TrimSpace(string(output))
			if processName == "app_process" {
				// 进程仍然在运行且是有效的scrcpy服务器进程
				s.logDebug("进程 %d 仍然在运行且有效，跳过终止", session.ProcessId)
				return nil
			}
		}
	}

	// 如果进程确实需要被终止，执行终止操作
	if session.ProcessId > 0 {
		// 1. 尝试使用adb shell kill命令终止
		killCmd := exec.Command("adb", "-s", session.DeviceId, "shell",
			fmt.Sprintf("kill -9 %d", session.ProcessId))
		_ = killCmd.Run() // 忽略错误，因为我们会在后面进行验证
	}

	// 2. 查找并终止所有scrcpy服务器进程
	// 使用更精确的grep模式避免误杀其他进程
	cmd := exec.Command("adb", "-s", session.DeviceId, "shell",
		"ps -ef | grep 'app_process.*com.genymobile.scrcpy.Server' | grep -v grep | awk '{print $2}' | xargs -r kill -9")

	if err := cmd.Run(); err != nil {
		// 如果没有找到进程或进程已经终止，通常会返回错误，可以忽略
		s.logDebug("终止服务器进程命令返回: %v", err)
	}

	// 3. 验证进程是否已经终止
	for i := 0; i < 3; i++ { // 最多尝试3次
		// 检查进程是否仍然存在
		pid, _ := s.getServerPid(ctx, session.DeviceId)
		if pid <= 0 {
			// 进程已终止
			return nil
		}

		// 等待一段时间再次检查
		time.Sleep(300 * time.Millisecond)
	}

	// 最后检查一次
	if pid, _ := s.getServerPid(ctx, session.DeviceId); pid > 0 {
		return gerror.Newf("无法终止服务器进程 (PID: %d)，多次尝试后仍然存在", pid)
	}

	return nil
}

// 设置ADB端口转发
func (s *sScrcpy) setupPortForward(ctx context.Context, session *StreamSession) (int, error) {
	// 1. 检查是否已有端口转发
	cmd := exec.Command("adb", "-s", session.DeviceId, "forward", "--list")
	output, err := cmd.Output()
	if err != nil {
		return 0, gerror.Wrap(err, "检查端口转发失败")
	}

	// 解析现有端口转发
	existingForwards := make(map[string]int)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			device := parts[0]
			localPort := parts[1]
			remotePort := parts[2]
			if strings.HasPrefix(localPort, "tcp:") && strings.HasPrefix(remotePort, "tcp:") {
				local, _ := strconv.Atoi(strings.TrimPrefix(localPort, "tcp:"))
				remote, _ := strconv.Atoi(strings.TrimPrefix(remotePort, "tcp:"))
				if remote == session.Port {
					existingForwards[device] = local
				}
			}
		}
	}

	// 2. 如果已有转发，检查是否可用
	if localPort, exists := existingForwards[session.DeviceId]; exists {
		s.logDebug("发现现有端口转发: 设备=%s, 本地端口=%d, 远程端口=%d",
			session.DeviceId, localPort, session.Port)

		// 测试端口是否可用
		if s.testConnection(ctx, session.DeviceId, localPort) {
			s.logDebug("现有端口转发可用，复用端口: %d", localPort)
			return localPort, nil
		}

		// 如果端口不可用，移除旧的转发
		s.logDebug("现有端口转发不可用，移除并重新创建")
		removeCmd := exec.Command("adb", "-s", session.DeviceId, "forward",
			fmt.Sprintf("--remove tcp:%d", localPort))
		_ = removeCmd.Run()
	}

	// 3. 创建新的端口转发
	localPort := 10000
	maxPort := 20000
	for localPort < maxPort {
		// 检查端口是否已被其他设备使用
		portInUse := false
		for device, port := range existingForwards {
			if port == localPort && device != session.DeviceId {
				portInUse = true
				break
			}
		}

		if !portInUse {
			// 尝试设置端口转发
			forwardCmd := exec.Command("adb", "-s", session.DeviceId, "forward",
				fmt.Sprintf("tcp:%d", localPort), fmt.Sprintf("tcp:%d", session.Port))

			if err := forwardCmd.Run(); err == nil {
				s.logDebug("创建新的端口转发: 设备=%s, 本地端口=%d, 远程端口=%d",
					session.DeviceId, localPort, session.Port)
				return localPort, nil
			}
		}

		// 尝试下一个端口
		localPort++
	}

	return 0, gerror.New("无法找到可用端口")
}

// 移除端口转发
func (s *sScrcpy) removePortForward(ctx context.Context, session *StreamSession) {
	// 尝试移除端口转发
	cmd := exec.Command("adb", "-s", session.DeviceId, "forward", "--remove",
		fmt.Sprintf("tcp:%d", session.LocalPort))

	// 忽略错误，仅记录日志
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to remove port forward: %v\n", err)
	}
}

// 检查设备连接状态
func (s *sScrcpy) checkDeviceStatus(ctx context.Context, deviceId string) error {
	cmd := exec.Command("adb", "devices")
	output, err := cmd.Output()
	if err != nil {
		return gerror.Wrap(err, "无法获取设备列表")
	}

	// 解析adb devices输出
	lines := strings.Split(string(output), "\n")
	deviceFound := false
	deviceStatus := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, deviceId) {
			deviceFound = true
			parts := strings.Fields(line)
			if len(parts) > 1 {
				deviceStatus = parts[1]
			}
			break
		}
	}

	if !deviceFound {
		return gerror.Newf("设备 %s 未连接", deviceId)
	}

	if deviceStatus != "device" {
		return gerror.Newf("设备 %s 状态异常: %s", deviceId, deviceStatus)
	}

	// 检查设备是否响应
	pingCmd := exec.Command("adb", "-s", deviceId, "shell", "echo OK")
	pingOutput, err := pingCmd.CombinedOutput()
	if err != nil || !strings.Contains(string(pingOutput), "OK") {
		return gerror.Newf("设备 %s 无响应", deviceId)
	}

	return nil
}

// 验证现有会话是否有效
func (s *sScrcpy) validateExistingSession(ctx context.Context, session *StreamSession) bool {
	// 1. 检查进程是否存在
	pid, err := s.getServerPid(ctx, session.DeviceId)
	if err != nil || pid <= 0 || pid != session.ProcessId {
		s.logDebug("会话验证失败: 进程不存在或已变更")
		return false
	}

	// 2. 检查本地端口转发是否有效
	checkCmd := exec.Command("adb", "-s", session.DeviceId, "forward", "--list")
	output, err := checkCmd.Output()
	if err != nil {
		s.logDebug("会话验证失败: 无法获取端口转发列表")
		return false
	}

	portForwardExists := false
	expectedForward := fmt.Sprintf("%s tcp:%d tcp:%d",
		session.DeviceId, session.LocalPort, session.Port)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == expectedForward {
			portForwardExists = true
			break
		}
	}

	if !portForwardExists {
		s.logDebug("会话验证失败: 端口转发不存在")
		return false
	}

	// 3. 检查本地端口是否可达
	if !s.isLocalPortReachable(session.LocalPort) {
		s.logDebug("会话验证失败: 本地端口不可达")
		return false
	}

	return true
}

// 检查本地端口是否可以连接
func (s *sScrcpy) isLocalPortReachable(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
