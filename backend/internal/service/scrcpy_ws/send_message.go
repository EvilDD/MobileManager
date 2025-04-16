package scrcpy_ws

import (
	"context"
	"encoding/binary"
	"net"
	"time"

	"github.com/gogf/gf/v2/os/glog"

	"backend/internal/model"
)

// sendVideoSettings 发送视频设置消息
func (s *ScrcpyService) sendVideoSettings(ctx context.Context, tcpConn net.Conn, deviceConn *model.DeviceConnection, settings model.VideoSettings) {
	// 创建视频设置消息
	buffer := make([]byte, 36) // 增加缓冲区大小到36字节，确保有足够空间

	// 设置消息类型
	buffer[0] = model.TYPE_CHANGE_STREAM_PARAMETERS

	// 写入比特率 (4字节)
	binary.BigEndian.PutUint32(buffer[1:5], uint32(settings.Bitrate))

	// 写入最大帧率 (4字节)
	binary.BigEndian.PutUint32(buffer[5:9], uint32(settings.MaxFps))

	// 写入I帧间隔 (1字节)
	buffer[9] = byte(settings.IFrameInterval)

	// 写入宽度 (2字节)
	binary.BigEndian.PutUint16(buffer[10:12], uint16(settings.Bounds.Width))

	// 写入高度 (2字节)
	binary.BigEndian.PutUint16(buffer[12:14], uint16(settings.Bounds.Height))

	// 写入裁剪区域 (8字节)
	// left, top, right, bottom 都设为0
	for i := 14; i < 22; i++ {
		buffer[i] = 0
	}

	// 写入是否发送帧元数据 (1字节)
	if settings.SendFrameMeta {
		buffer[22] = 1
	} else {
		buffer[22] = 0
	}

	// 写入锁定视频方向 (1字节)
	buffer[23] = byte(settings.LockedVideoOrientation)

	// 写入显示ID (4字节)
	binary.BigEndian.PutUint32(buffer[24:28], uint32(settings.DisplayId))

	// 写入编解码器选项长度 (4字节)
	binary.BigEndian.PutUint32(buffer[28:32], 0)

	// 写入编码器名称长度 (4字节)
	binary.BigEndian.PutUint32(buffer[32:36], 0)

	// 发送消息
	_, err := tcpConn.Write(buffer)
	if err != nil {
		glog.Error(ctx, "发送视频设置失败:", err)
		return
	}

	deviceConn.VideoSettingsSent = true
	glog.Info(ctx, "视频设置已发送",
		"比特率:", settings.Bitrate,
		"最大帧率:", settings.MaxFps,
		"分辨率:", settings.Bounds.Width, "x", settings.Bounds.Height,
		"锁定方向:", settings.LockedVideoOrientation,
		"发送帧元数据:", settings.SendFrameMeta)
}

// sendTouchEvent 发送触摸事件
func (s *ScrcpyService) sendTouchEvent(ctx context.Context, tcpConn net.Conn, deviceConn *model.DeviceConnection, action, x, y int) {
	// 创建触摸事件消息
	buffer := make([]byte, 30) // 固定大小为30字节

	// 设置消息类型
	buffer[0] = model.TYPE_INJECT_TOUCH_EVENT

	// 设置动作类型
	buffer[1] = byte(action)

	// 写入pointerId高32位和低32位 (8字节)

	// 写入X坐标 (4字节)
	binary.BigEndian.PutUint32(buffer[10:14], uint32(x))

	// 写入Y坐标 (4字节)
	binary.BigEndian.PutUint32(buffer[14:18], uint32(y))

	// 写入屏幕宽度 (2字节) - 使用实际视频宽度
	if deviceConn.VideoWidth > 0 {
		binary.BigEndian.PutUint16(buffer[18:20], uint16(deviceConn.VideoWidth))
	} else {
		binary.BigEndian.PutUint16(buffer[18:20], uint16(deviceConn.ScreenWidth))
	}

	// 写入屏幕高度 (2字节) - 使用实际视频高度
	if deviceConn.VideoHeight > 0 {
		binary.BigEndian.PutUint16(buffer[20:22], uint16(deviceConn.VideoHeight))
	} else {
		binary.BigEndian.PutUint16(buffer[20:22], uint16(deviceConn.ScreenHeight))
	}

	// 写入压力值 (2字节)
	pressure := uint16(0)
	if action == model.ACTION_DOWN {
		pressure = 0xFFFF
	}
	binary.BigEndian.PutUint16(buffer[22:24], pressure)

	// 写入按钮值 (4字节)
	binary.BigEndian.PutUint32(buffer[24:28], model.BUTTON_PRIMARY)

	// 发送消息
	_, err := tcpConn.Write(buffer)
	if err != nil {
		glog.Error(ctx, "发送触摸事件失败:", err)
		return
	}

	actionName := "未知"
	switch action {
	case model.ACTION_DOWN:
		actionName = "DOWN"
	case model.ACTION_UP:
		actionName = "UP"
	case model.ACTION_MOVE:
		actionName = "MOVE"
	}

	glog.Info(ctx, "触摸事件已发送",
		"类型:", actionName,
		"坐标:", x, y,
		"视频尺寸:", deviceConn.VideoWidth, "x", deviceConn.VideoHeight)
}

// sendSwipeEvent 发送滑动事件
func (s *ScrcpyService) sendSwipeEvent(ctx context.Context, tcpConn net.Conn, deviceConn *model.DeviceConnection, event model.SwipeEvent) {
	glog.Info(ctx, "开始滑动:", "起点:", event.StartX, event.StartY, "终点:", event.EndX, event.EndY)

	// 发送按下事件
	s.sendTouchEvent(ctx, tcpConn, deviceConn, model.ACTION_DOWN, event.StartX, event.StartY)
	time.Sleep(50 * time.Millisecond) // 短暂延迟

	// 计算每一步的移动距离
	steps := event.Steps
	if steps <= 0 {
		steps = 10 // 默认步数
	}

	duration := float64(event.Duration)
	if duration <= 0 {
		duration = 500 // 默认500毫秒
	}

	xStep := float64(event.EndX-event.StartX) / float64(steps)
	yStep := float64(event.EndY-event.StartY) / float64(steps)
	stepDelay := duration / float64(steps)

	// 发送移动事件
	for i := 1; i <= steps; i++ {
		currentX := int(float64(event.StartX) + xStep*float64(i))
		currentY := int(float64(event.StartY) + yStep*float64(i))
		s.sendTouchEvent(ctx, tcpConn, deviceConn, model.ACTION_MOVE, currentX, currentY)
		time.Sleep(time.Duration(stepDelay) * time.Millisecond)
	}

	// 发送抬起事件
	s.sendTouchEvent(ctx, tcpConn, deviceConn, model.ACTION_UP, event.EndX, event.EndY)
	glog.Info(ctx, "滑动事件已完成")
}

// sendClickEvent 发送点击事件
func (s *ScrcpyService) sendClickEvent(ctx context.Context, tcpConn net.Conn, deviceConn *model.DeviceConnection, event model.ClickEvent) {
	glog.Info(ctx, "开始点击:", "坐标:", event.X, event.Y)

	// 发送按下事件
	s.sendTouchEvent(ctx, tcpConn, deviceConn, model.ACTION_DOWN, event.X, event.Y)

	// 点击持续时间
	duration := event.Duration
	if duration <= 0 {
		duration = 100 // 默认100毫秒的点击持续时间
	}

	// 等待指定的点击持续时间
	time.Sleep(time.Duration(duration) * time.Millisecond)

	// 发送抬起事件
	s.sendTouchEvent(ctx, tcpConn, deviceConn, model.ACTION_UP, event.X, event.Y)
	glog.Info(ctx, "点击事件已完成")
}
