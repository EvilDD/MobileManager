package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	v1 "backend/api/scrcpy/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
)

type sScrcpy struct {
	sessions      map[string]*StreamSession
	sessionsLock  sync.RWMutex
	serverJarPath string
}

type StreamSession struct {
	DeviceId  string
	Port      int
	ProcessId int
}

func init() {
	localScrcpy = New()
}

func New() *sScrcpy {
	return &sScrcpy{
		sessions:      make(map[string]*StreamSession),
		serverJarPath: "resource/scrcpy/scrcpy-server.jar",
	}
}

// StartStream 启动设备流服务
func (s *sScrcpy) StartStream(ctx context.Context, req *v1.StartStreamReq) (res *v1.StartStreamRes, err error) {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	// 检查会话是否已存在
	if session, exists := s.sessions[req.DeviceId]; exists {
		return &v1.StartStreamRes{
			Port: session.Port,
			Url:  fmt.Sprintf("ws://localhost:%d", session.Port),
		}, nil
	}

	// 创建新会话
	session := &StreamSession{
		DeviceId: req.DeviceId,
		Port:     8886, // 可以改为动态分配端口
	}

	// 推送服务器文件
	if err := s.pushServer(ctx, req.DeviceId); err != nil {
		return nil, err
	}

	// 启动服务器
	if err := s.startServer(ctx, session); err != nil {
		return nil, err
	}

	s.sessions[req.DeviceId] = session

	return &v1.StartStreamRes{
		Port: session.Port,
		Url:  fmt.Sprintf("ws://localhost:%d", session.Port),
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

	// 停止服务器进程
	if err := s.killServer(ctx, session); err != nil {
		return nil, err
	}

	delete(s.sessions, req.DeviceId)
	return &v1.StopStreamRes{}, nil
}

// 内部辅助方法W
func (s *sScrcpy) pushServer(ctx context.Context, deviceId string) error {
	jarPath := gfile.Join(gfile.MainPkgPath(), s.serverJarPath)
	if !gfile.Exists(jarPath) {
		return gerror.Newf("scrcpy-server.jar not found at: %s", jarPath)
	}

	cmd := exec.Command("adb", "-s", deviceId, "push", jarPath, "/data/local/tmp/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return gerror.Wrap(err, "Failed to push scrcpy-server.jar")
	}

	return nil
}

func (s *sScrcpy) startServer(ctx context.Context, session *StreamSession) error {
	// 启动服务器命令
	cmd := exec.Command("adb", "-s", session.DeviceId, "shell",
		fmt.Sprintf("CLASSPATH=/data/local/tmp/scrcpy-server.jar nohup app_process / com.genymobile.scrcpy.Server 1.19-ws6 web ERROR  %d true 2>&1 > /dev/null",
			session.Port))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return gerror.Wrap(err, "Failed to start scrcpy server")
	}

	session.ProcessId = cmd.Process.Pid

	// 启动后台进程监控
	go func() {
		cmd.Wait()
		s.sessionsLock.Lock()
		delete(s.sessions, session.DeviceId)
		s.sessionsLock.Unlock()
	}()

	return nil
}

func (s *sScrcpy) killServer(ctx context.Context, session *StreamSession) error {
	// 终止服务器进程
	cmd := exec.Command("adb", "-s", session.DeviceId, "shell",
		fmt.Sprintf("pkill -f 'app_process.*com.genymobile.scrcpy.Server'"))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return gerror.Wrap(err, "Failed to kill scrcpy server")
	}

	return nil
}
