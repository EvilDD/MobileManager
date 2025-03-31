package adb

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ADB 工具类接口
type IAdb interface {
	// Connect 连接设备
	Connect(deviceId string) error
	// ExecuteCommand 执行 ADB 命令
	ExecuteCommand(deviceId string, args ...string) (string, error)
	// PullFile 从设备拉取文件
	PullFile(deviceId string, devicePath string, localPath string) error
	// RemoveDeviceFile 删除设备上的文件
	RemoveDeviceFile(deviceId string, path string) error
	// Screencap 设备截图
	Screencap(deviceId string, savePath string) error
}

type adbService struct{}

var defaultAdb IAdb

func init() {
	defaultAdb = &adbService{}
}

// Default 获取默认的 ADB 工具实例
func Default() IAdb {
	return defaultAdb
}

// Connect 连接设备
func (s *adbService) Connect(deviceId string) error {
	cmd := exec.Command("adb", "connect", deviceId)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("连接设备失败: %v, output: %s", err, string(output))
	}
	return nil
}

// ExecuteCommand 执行 ADB 命令
func (s *adbService) ExecuteCommand(deviceId string, args ...string) (string, error) {
	cmdArgs := append([]string{"-s", deviceId}, args...)
	cmd := exec.Command("adb", cmdArgs...)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("执行ADB命令失败: %v", err)
	}

	return out.String(), nil
}

// PullFile 从设备拉取文件
func (s *adbService) PullFile(deviceId string, devicePath string, localPath string) error {
	_, err := s.ExecuteCommand(deviceId, "pull", devicePath, localPath)
	if err != nil {
		return fmt.Errorf("拉取文件失败: %v", err)
	}
	return nil
}

// RemoveDeviceFile 删除设备上的文件
func (s *adbService) RemoveDeviceFile(deviceId string, path string) error {
	_, err := s.ExecuteCommand(deviceId, "shell", "rm", path)
	if err != nil {
		return fmt.Errorf("删除设备文件失败: %v", err)
	}
	return nil
}

// Screencap 设备截图
func (s *adbService) Screencap(deviceId string, savePath string) error {
	_, err := s.ExecuteCommand(deviceId, "shell", "screencap", "-p", savePath)
	if err != nil {
		return fmt.Errorf("设备截图失败: %v", err)
	}
	return nil
}

// 执行ADB命令的基础函数
func executeAdbCommand(args ...string) (string, error) {
	cmd := exec.Command("adb", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// 执行针对特定设备的ADB命令
func executeDeviceAdbCommand(deviceId string, args ...string) (string, error) {
	cmdArgs := append([]string{"-s", deviceId}, args...)
	return executeAdbCommand(cmdArgs...)
}

// InstallApp 安装应用到设备
func InstallApp(deviceId, apkPath string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "install", "-r", apkPath)
}

// UninstallApp 从设备卸载应用
func UninstallApp(deviceId, packageName string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "uninstall", packageName)
}

// StartApp 启动应用
func StartApp(deviceId, packageName string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "shell", "monkey", "-p", packageName, "-c", "android.intent.category.LAUNCHER", "1")
}

// GetInstalledApps 获取设备上已安装的应用列表
func GetInstalledApps(deviceId string) ([]string, error) {
	output, err := executeDeviceAdbCommand(deviceId, "shell", "pm", "list", "packages")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var packageNames []string

	for _, line := range lines {
		if strings.HasPrefix(line, "package:") {
			packageName := strings.TrimSpace(strings.TrimPrefix(line, "package:"))
			packageNames = append(packageNames, packageName)
		}
	}

	return packageNames, nil
}

// GetAppInfo 获取应用信息
func GetAppInfo(deviceId, packageName string) (map[string]string, error) {
	output, err := executeDeviceAdbCommand(deviceId, "shell", "dumpsys", "package", packageName)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "versionName=") {
			info["versionName"] = strings.TrimPrefix(line, "versionName=")
		} else if strings.HasPrefix(line, "versionCode=") {
			info["versionCode"] = strings.TrimPrefix(line, "versionCode=")
		} else if strings.HasPrefix(line, "firstInstallTime=") {
			info["firstInstallTime"] = strings.TrimPrefix(line, "firstInstallTime=")
		} else if strings.HasPrefix(line, "lastUpdateTime=") {
			info["lastUpdateTime"] = strings.TrimPrefix(line, "lastUpdateTime=")
		}
	}

	return info, nil
}

// StopApp 停止应用
func StopApp(deviceId, packageName string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "shell", "am", "force-stop", packageName)
}

// ClearAppData 清除应用数据
func ClearAppData(deviceId, packageName string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "shell", "pm", "clear", packageName)
}

// GoToHome 回到主菜单
func GoToHome(deviceId string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "shell", "input", "keyevent", "3")
}

// KillAllBackgroundApps 清除所有后台应用
func KillAllBackgroundApps(deviceId string) (string, error) {
	return executeDeviceAdbCommand(deviceId, "shell", "am", "kill-all")
}
