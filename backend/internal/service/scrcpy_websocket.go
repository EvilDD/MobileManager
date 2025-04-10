package service

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// DevicePortMap 保存设备ID和转发端口的映射关系
type DevicePortMap struct {
	UdId        string    // 设备ID
	Port        int       // 转发端口
	CreatedTime time.Time // 创建时间
}

// ScrcpyWebsocketService 管理scrcpy WebSocket服务
type ScrcpyWebsocketService struct {
	// 设备ID到端口的映射
	devicePortMap sync.Map
	// 端口计数器，用于分配新端口
	portCounter int
	// 端口基数，实际端口 = 端口基数 + 端口计数器
	portBase int
	// 保护端口计数器的锁
	portMutex sync.Mutex
}

var (
	// scrcpyWsServiceInstance 单例实例
	scrcpyWsServiceInstance *ScrcpyWebsocketService
	// 保护单例初始化的锁
	scrcpyWsServiceOnce sync.Once
)

// GetScrcpyWebsocketService 获取scrcpy WebSocket服务实例
func GetScrcpyWebsocketService() *ScrcpyWebsocketService {
	scrcpyWsServiceOnce.Do(func() {
		scrcpyWsServiceInstance = &ScrcpyWebsocketService{
			portBase: 28000, // 使用28000作为端口基数
		}

		// 启动定期清理任务
		go scrcpyWsServiceInstance.startCleanupTask()
	})
	return scrcpyWsServiceInstance
}

// RegisterDevicePort 注册设备与端口的映射关系
func (s *ScrcpyWebsocketService) RegisterDevicePort(ctx context.Context, udid string, port int) {
	// 记录设备和端口的映射关系
	s.devicePortMap.Store(udid, &DevicePortMap{
		UdId:        udid,
		Port:        port,
		CreatedTime: time.Now(),
	})
	glog.Info(ctx, "已注册设备端口映射", "设备ID:", udid, "端口:", port)
}

// GetDevicePort 获取设备的转发端口
func (s *ScrcpyWebsocketService) GetDevicePort(ctx context.Context, udid string) (int, error) {
	// 查找设备端口映射
	if value, ok := s.devicePortMap.Load(udid); ok {
		mapping := value.(*DevicePortMap)
		glog.Debug(ctx, "找到设备端口映射", "设备ID:", udid, "端口:", mapping.Port)
		return mapping.Port, nil
	}

	// 没有找到映射
	return 0, gerror.Newf("未找到设备 %s 的端口映射", udid)
}

// AllocateNewPort 分配一个新的端口
func (s *ScrcpyWebsocketService) AllocateNewPort(ctx context.Context) int {
	s.portMutex.Lock()
	defer s.portMutex.Unlock()

	s.portCounter++
	newPort := s.portBase + s.portCounter
	glog.Debug(ctx, "分配新端口", "端口:", newPort)
	return newPort
}

// RemoveDevicePort 移除设备端口映射
func (s *ScrcpyWebsocketService) RemoveDevicePort(ctx context.Context, udid string) {
	// 查找设备端口映射
	if value, ok := s.devicePortMap.Load(udid); ok {
		mapping := value.(*DevicePortMap)
		glog.Info(ctx, "移除设备端口映射", "设备ID:", udid, "端口:", mapping.Port)
	}

	// 删除映射
	s.devicePortMap.Delete(udid)
}

// ListDevicePorts 列出所有设备端口映射
func (s *ScrcpyWebsocketService) ListDevicePorts(ctx context.Context) []map[string]interface{} {
	var result []map[string]interface{}

	s.devicePortMap.Range(func(key, value interface{}) bool {
		mapping := value.(*DevicePortMap)
		result = append(result, map[string]interface{}{
			"udid":    mapping.UdId,
			"port":    mapping.Port,
			"created": mapping.CreatedTime.Format("2006-01-02 15:04:05"),
		})
		return true
	})

	return result
}

// cleanupOldMappings 清理旧的映射
func (s *ScrcpyWebsocketService) cleanupOldMappings() {
	ctx := gctx.New()
	glog.Debug(ctx, "开始清理旧的设备端口映射")

	now := time.Now()
	deletedCount := 0

	s.devicePortMap.Range(func(key, value interface{}) bool {
		mapping := value.(*DevicePortMap)
		// 如果映射超过12小时，删除它
		if now.Sub(mapping.CreatedTime) > 12*time.Hour {
			s.devicePortMap.Delete(key)
			deletedCount++
			glog.Debug(ctx, "删除过期的设备端口映射", "设备ID:", mapping.UdId, "端口:", mapping.Port)
		}
		return true
	})

	if deletedCount > 0 {
		glog.Info(ctx, "已清理过期的设备端口映射", "数量:", deletedCount)
	}
}

// startCleanupTask 启动定期清理任务
func (s *ScrcpyWebsocketService) startCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := gctx.New()
		g.Try(ctx, func(ctx context.Context) {
			s.cleanupOldMappings()
		})
	}
}
