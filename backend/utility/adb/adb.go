package adb

import (
	"os/exec"
	"strings"
)

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
