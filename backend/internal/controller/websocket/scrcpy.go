package websocket

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"
)

// 常量定义
const (
	// 消息魔法字节
	MAGIC_BYTES_INITIAL = "scrcpy_initial"
	MAGIC_BYTES_MESSAGE = "scrcpy_message"

	// 控制消息类型
	TYPE_INJECT_KEYCODE            = 0
	TYPE_INJECT_TEXT               = 1
	TYPE_INJECT_TOUCH_EVENT        = 2
	TYPE_INJECT_SCROLL_EVENT       = 3
	TYPE_BACK_OR_SCREEN_ON         = 4
	TYPE_EXPAND_NOTIFICATION_PANEL = 5
	TYPE_EXPAND_SETTINGS_PANEL     = 6
	TYPE_COLLAPSE_PANELS           = 7
	TYPE_GET_CLIPBOARD             = 8
	TYPE_SET_CLIPBOARD             = 9
	TYPE_SET_SCREEN_POWER_MODE     = 10
	TYPE_ROTATE_DEVICE             = 11
	TYPE_CHANGE_STREAM_PARAMETERS  = 101

	// 触摸事件动作
	ACTION_DOWN = 0
	ACTION_UP   = 1
	ACTION_MOVE = 2

	// 按钮标识
	BUTTON_PRIMARY = 1 << 0 // 左键
)

// ScrcpyController 处理scrcpy的WebSocket连接和消息转发
type ScrcpyController struct {
	// 升级HTTP连接到WebSocket的upgrader
	upgrader websocket.Upgrader
	// 设备ID到TCP连接的映射
	deviceConnections sync.Map
	// 保护deviceConnections的互斥锁
	connMutex sync.RWMutex
}

// DeviceConnection 存储设备连接信息
type DeviceConnection struct {
	UdId              string    // 设备ID
	Port              int       // ADB转发的端口
	Conn              net.Conn  // 到ADB转发端口的TCP连接
	LastUsed          time.Time // 最后使用时间
	ScreenWidth       int       // 屏幕宽度
	ScreenHeight      int       // 屏幕高度
	VideoWidth        int       // 视频实际宽度
	VideoHeight       int       // 视频实际高度
	ClientId          int       // 客户端ID
	HasInitInfo       bool      // 是否已接收初始化信息
	VideoSettingsSent bool      // 是否已发送视频设置
}

// TouchEvent 触摸事件结构
type TouchEvent struct {
	Action int `json:"action"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// SwipeEvent 滑动事件结构
type SwipeEvent struct {
	StartX   int `json:"startX"`
	StartY   int `json:"startY"`
	EndX     int `json:"endX"`
	EndY     int `json:"endY"`
	Duration int `json:"duration"`
	Steps    int `json:"steps"`
}

// VideoSettings 视频设置结构
type VideoSettings struct {
	Bitrate        int `json:"bitrate"`
	MaxFps         int `json:"maxFps"`
	IFrameInterval int `json:"iFrameInterval"`
	Width          int `json:"width"`
	Height         int `json:"height"`
}

// NewScrcpyController 创建一个新的ScrcpyController
func NewScrcpyController() *ScrcpyController {
	return &ScrcpyController{
		upgrader: websocket.Upgrader{
			// 允许所有CORS请求
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Handler 处理WebSocket请求
func (c *ScrcpyController) Handler(r *ghttp.Request) {
	ctx := r.GetCtx()
	glog.Info(ctx, "收到WebSocket连接请求", "URL:", r.URL.String(), "Headers:", r.Header)

	// 解析设备ID
	udid := r.Get("udid").String()
	glog.Info(ctx, "解析设备ID参数", "udid:", udid)
	if udid == "" {
		glog.Error(ctx, "缺少设备ID参数")
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "缺少设备ID参数"})
		return
	}

	// 解析ADB转发端口
	port := r.Get("port").Int()
	glog.Info(ctx, "解析端口参数", "port:", port)
	if port <= 0 {
		glog.Error(ctx, "无效的端口参数")
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": http.StatusBadRequest, "message": "无效的端口参数"})
		return
	}

	// 升级HTTP连接到WebSocket
	wsConn, err := c.upgrader.Upgrade(r.Response.Writer, r.Request, nil)
	if err != nil {
		glog.Error(ctx, "升级WebSocket连接失败:", err)
		return
	}
	defer wsConn.Close()

	// 预先检查ADB端口转发是否正确设置
	if err := c.checkPortForward(ctx, udid, port); err != nil {
		glog.Error(ctx, "端口转发检查失败:", err, "设备ID:", udid, "端口:", port)

		// 向客户端发送结构化的错误消息
		errorMsg := map[string]interface{}{
			"type":    "error",
			"code":    "PORT_FORWARD_NOT_FOUND",
			"message": fmt.Sprintf("端口转发检查失败: %v", err),
			"data": map[string]interface{}{
				"deviceId":    udid,
				"port":        port,
				"errorDetail": err.Error(),
			},
		}

		errorJSON, _ := json.Marshal(errorMsg)
		wsConn.WriteMessage(websocket.TextMessage, errorJSON)

		return
	}

	// 连接到ADB转发的端口
	tcpConn, err := c.connectToDevice(ctx, udid, port)
	if err != nil {
		glog.Error(ctx, "连接到设备失败:", err, "设备ID:", udid, "端口:", port)

		// 向客户端发送结构化的错误消息
		errorMsg := map[string]interface{}{
			"type":    "error",
			"code":    "DEVICE_CONNECTION_FAILED",
			"message": fmt.Sprintf("连接到设备失败: %v", err),
			"data": map[string]interface{}{
				"deviceId":    udid,
				"port":        port,
				"errorDetail": err.Error(),
			},
		}

		errorJSON, _ := json.Marshal(errorMsg)
		wsConn.WriteMessage(websocket.TextMessage, errorJSON)

		return
	}
	defer tcpConn.Close()

	// 记录连接信息
	deviceConn := &DeviceConnection{
		UdId:              udid,
		Port:              port,
		Conn:              tcpConn,
		LastUsed:          time.Now(),
		ScreenWidth:       0,
		ScreenHeight:      0,
		VideoWidth:        0,
		VideoHeight:       0,
		ClientId:          -1,
		HasInitInfo:       false,
		VideoSettingsSent: false,
	}
	c.deviceConnections.Store(udid, deviceConn)
	defer c.deviceConnections.Delete(udid)

	glog.Info(ctx, "WebSocket连接已建立", "设备ID:", udid, "端口:", port)

	// 向客户端发送连接成功的消息
	successMsg := map[string]interface{}{
		"type": "connected",
		"data": map[string]interface{}{
			"deviceId":  udid,
			"port":      port,
			"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
		},
	}
	successJSON, _ := json.Marshal(successMsg)
	wsConn.WriteMessage(websocket.TextMessage, successJSON)

	// 创建通道用于同步goroutine
	done := make(chan struct{})

	// 从WebSocket接收消息并转发到TCP连接
	go func() {
		defer close(done)
		for {
			// 读取WebSocket消息
			messageType, message, err := wsConn.ReadMessage()
			if err != nil {
				glog.Error(ctx, "读取WebSocket消息失败:", err)
				return
			}

			// 更新最后使用时间
			deviceConn.LastUsed = time.Now()

			// 处理JSON命令消息
			if messageType == websocket.TextMessage {
				c.handleCommandMessage(ctx, wsConn, tcpConn, deviceConn, message)
				continue
			}

			// 处理二进制消息
			if messageType == websocket.BinaryMessage {
				// 直接转发到TCP连接
				_, err = tcpConn.Write(message)
				if err != nil {
					glog.Error(ctx, "向TCP连接写入数据失败:", err)
					return
				}
				glog.Debug(ctx, "转发到设备的消息大小:", len(message))
			}
		}
	}()

	// 从TCP连接接收消息并转发到WebSocket
	go func() {
		buffer := make([]byte, 32*1024) // 32KB的缓冲区
		for {
			// 从TCP连接读取数据
			n, err := tcpConn.Read(buffer)
			if err != nil {
				glog.Error(ctx, "从TCP连接读取数据失败:", err)

				// 向客户端发送连接断开的错误消息
				disconnectMsg := map[string]interface{}{
					"type":    "disconnected",
					"code":    "DEVICE_CONNECTION_CLOSED",
					"message": fmt.Sprintf("设备连接已断开: %v", err),
					"data": map[string]interface{}{
						"deviceId":    udid,
						"errorDetail": err.Error(),
					},
				}
				disconnectJSON, _ := json.Marshal(disconnectMsg)
				wsConn.WriteMessage(websocket.TextMessage, disconnectJSON)

				select {
				case <-done:
					return
				default:
					close(done)
					return
				}
			}

			// 更新最后使用时间
			deviceConn.LastUsed = time.Now()

			// 获取接收到的数据
			received := buffer[:n]

			// 处理特殊消息类型
			if c.handleSpecialMessages(ctx, wsConn, deviceConn, received) {
				continue
			}

			// 转发到WebSocket
			err = wsConn.WriteMessage(websocket.BinaryMessage, received)
			if err != nil {
				glog.Error(ctx, "向WebSocket写入数据失败:", err)
				return
			}
			// glog.Debug(ctx, "从设备转发的消息大小:", n)
		}
	}()

	// 等待直到连接关闭
	<-done
	glog.Info(ctx, "WebSocket连接已关闭", "设备ID:", udid)
}

// handleCommandMessage 处理从客户端发来的JSON命令消息
func (c *ScrcpyController) handleCommandMessage(ctx context.Context, wsConn *websocket.Conn, tcpConn net.Conn, deviceConn *DeviceConnection, message []byte) {
	var command map[string]interface{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		glog.Error(ctx, "解析JSON命令失败:", err)
		return
	}

	// 获取命令类型
	cmdType, ok := command["type"].(string)
	if !ok {
		glog.Error(ctx, "无效的命令类型")
		return
	}

	switch cmdType {
	case "touch":
		// 处理触摸事件
		var touchEvent TouchEvent
		if data, ok := command["data"].(map[string]interface{}); ok {
			touchEvent.Action = int(data["action"].(float64))
			touchEvent.X = int(data["x"].(float64))
			touchEvent.Y = int(data["y"].(float64))
			c.sendTouchEvent(ctx, tcpConn, deviceConn, touchEvent.Action, touchEvent.X, touchEvent.Y)
		}
	case "swipe":
		// 处理滑动事件
		var swipeEvent SwipeEvent
		if data, ok := command["data"].(map[string]interface{}); ok {
			swipeEvent.StartX = int(data["startX"].(float64))
			swipeEvent.StartY = int(data["startY"].(float64))
			swipeEvent.EndX = int(data["endX"].(float64))
			swipeEvent.EndY = int(data["endY"].(float64))
			swipeEvent.Duration = int(data["duration"].(float64))
			swipeEvent.Steps = int(data["steps"].(float64))
			c.sendSwipeEvent(ctx, tcpConn, deviceConn, swipeEvent)
		}
	case "videoSettings":
		// 处理视频设置
		var videoSettings VideoSettings
		if data, ok := command["data"].(map[string]interface{}); ok {
			videoSettings.Bitrate = int(data["bitrate"].(float64))
			videoSettings.MaxFps = int(data["maxFps"].(float64))
			videoSettings.IFrameInterval = int(data["iFrameInterval"].(float64))
			videoSettings.Width = int(data["width"].(float64))
			videoSettings.Height = int(data["height"].(float64))
			c.sendVideoSettings(ctx, tcpConn, deviceConn, videoSettings)
		}
	default:
		glog.Warning(ctx, "未知命令类型:", cmdType)
	}
}

// handleSpecialMessages 处理特殊消息类型
func (c *ScrcpyController) handleSpecialMessages(ctx context.Context, wsConn *websocket.Conn, deviceConn *DeviceConnection, data []byte) bool {
	// 检查是否是初始化消息
	if len(data) > len(MAGIC_BYTES_INITIAL) && strings.HasPrefix(string(data), MAGIC_BYTES_INITIAL) {
		c.handleInitialInfo(ctx, wsConn, deviceConn, data)
		return true
	}

	// 检查是否是设备消息
	if len(data) > len(MAGIC_BYTES_MESSAGE) && strings.HasPrefix(string(data), MAGIC_BYTES_MESSAGE) {
		c.handleDeviceMessage(ctx, wsConn, deviceConn, data)
		return true
	}

	// 检查是否是 SPS 数据
	if len(data) >= 5 {
		nalType := data[4] & 0x1F
		if nalType == 7 { // NAL type 7 表示 SPS
			spsInfo, err := parseSPS(data)
			if err != nil {
				glog.Error(ctx, "解析 SPS 失败", "error", err)
				return false
			}

			// 记录原始尺寸
			originalWidth := deviceConn.ScreenWidth
			originalHeight := deviceConn.ScreenHeight

			// 计算实际视频尺寸
			width := (spsInfo.PicWidthInMbsMinus1 + 1) * 16
			if spsInfo.FrameCropLeftOffset > 0 || spsInfo.FrameCropRightOffset > 0 {
				width -= (spsInfo.FrameCropLeftOffset + spsInfo.FrameCropRightOffset) * 2
			}

			height := (spsInfo.PicHeightInMapUnitsMinus1 + 1) * 16
			if spsInfo.FrameMbsOnlyFlag == 0 {
				height *= 2
			}
			if spsInfo.FrameCropTopOffset > 0 || spsInfo.FrameCropBottomOffset > 0 {
				cropMult := uint(2)
				if spsInfo.FrameMbsOnlyFlag > 0 {
					cropMult = 1
				}
				height -= (spsInfo.FrameCropTopOffset + spsInfo.FrameCropBottomOffset) * cropMult
			}

			// 应用 SAR
			if spsInfo.Sar[0] != 0 && spsInfo.Sar[1] != 0 {
				width = uint(float64(width) * float64(spsInfo.Sar[0]) / float64(spsInfo.Sar[1]))
			}

			// 更新设备连接的视频尺寸
			deviceConn.VideoWidth = int(width)
			deviceConn.VideoHeight = int(height)

			// 构建编解码器字符串
			codec := fmt.Sprintf("avc1.%02X%02X%02X",
				spsInfo.ProfileIdc,
				spsInfo.ConstraintSetFlags,
				spsInfo.LevelIdc)

			glog.Info(ctx, "从视频帧解析到实际编码尺寸",
				"原尺寸", fmt.Sprintf("%d x %d", originalWidth, originalHeight),
				"新尺寸", fmt.Sprintf("%d x %d", deviceConn.VideoWidth, deviceConn.VideoHeight),
				"编解码器", codec)

			// 向客户端发送视频尺寸更新消息
			sizeUpdateMsg := map[string]interface{}{
				"type": "videoSize",
				"data": map[string]interface{}{
					"width":  width,
					"height": height,
					"codec":  codec,
				},
			}
			msgJSON, _ := json.Marshal(sizeUpdateMsg)
			wsConn.WriteMessage(websocket.TextMessage, msgJSON)
		}
	}

	return false
}

// handleInitialInfo 处理初始化信息
func (c *ScrcpyController) handleInitialInfo(ctx context.Context, wsConn *websocket.Conn, deviceConn *DeviceConnection, data []byte) {
	glog.Info(ctx, "处理初始化信息...")

	// 直接转发初始化信息到客户端
	err := wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		glog.Error(ctx, "转发初始化信息失败:", err)
		return
	}

	// 尝试解析屏幕尺寸和客户端ID，实际应用中可能需要更复杂的解析逻辑
	offset := len(MAGIC_BYTES_INITIAL)

	if len(data) > offset+100 { // 简化版，实际需要根据结构确定正确的偏移量
		// 设备名称占64字节，跳过
		offset += 64

		// 解析显示数量
		if offset+4 <= len(data) {
			displaysCount := int(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4

			if displaysCount > 0 && offset+24 <= len(data) {
				// 从第一个显示信息中提取宽高
				// 假设DisplayInfo结构第4-8字节是宽度，8-12字节是高度
				width := int(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
				height := int(binary.BigEndian.Uint32(data[offset+8 : offset+12]))

				deviceConn.ScreenWidth = width
				deviceConn.ScreenHeight = height
				glog.Info(ctx, "解析到屏幕尺寸:", "宽:", width, "高:", height)
			}
		}
	}

	deviceConn.HasInitInfo = true

	// 在初始化信息处理完成后发送视频设置（如果前端请求）
	if !deviceConn.VideoSettingsSent {
		// 使用默认值发送视频设置
		c.sendVideoSettings(ctx, deviceConn.Conn, deviceConn, VideoSettings{
			Bitrate:        50000,
			MaxFps:         24,
			IFrameInterval: 5,
			Width:          540,
			Height:         960,
		})
	}
}

// handleDeviceMessage 处理设备消息
func (c *ScrcpyController) handleDeviceMessage(ctx context.Context, wsConn *websocket.Conn, deviceConn *DeviceConnection, data []byte) {
	glog.Info(ctx, "处理设备消息...")

	// 直接转发设备消息到客户端
	err := wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		glog.Error(ctx, "转发设备消息失败:", err)
		return
	}

	// 解析消息类型，后续可以添加特定消息类型的处理
	if len(data) > len(MAGIC_BYTES_MESSAGE) {
		msgType := data[len(MAGIC_BYTES_MESSAGE)]
		glog.Debug(ctx, "设备消息类型:", msgType)
	}
}

// sendVideoSettings 发送视频设置消息
func (c *ScrcpyController) sendVideoSettings(ctx context.Context, tcpConn net.Conn, deviceConn *DeviceConnection, settings VideoSettings) {
	// 创建视频设置消息
	buffer := make([]byte, 36) // 增加缓冲区大小到36字节，确保有足够空间

	// 设置消息类型
	buffer[0] = TYPE_CHANGE_STREAM_PARAMETERS

	// 写入比特率 (4字节)
	binary.BigEndian.PutUint32(buffer[1:5], uint32(settings.Bitrate))

	// 写入最大帧率 (4字节)
	binary.BigEndian.PutUint32(buffer[5:9], uint32(settings.MaxFps))

	// 写入I帧间隔 (1字节)
	buffer[9] = byte(settings.IFrameInterval)

	// 写入宽度 (2字节)
	binary.BigEndian.PutUint16(buffer[10:12], uint16(settings.Width))

	// 写入高度 (2字节)
	binary.BigEndian.PutUint16(buffer[12:14], uint16(settings.Height))

	// 写入裁剪区域 (8字节)
	// left, top, right, bottom 都设为0
	for i := 14; i < 22; i++ {
		buffer[i] = 0
	}

	// 写入是否发送帧元数据 (1字节)
	buffer[22] = 0

	// 写入锁定视频方向 (1字节)
	buffer[23] = 0xFF // -1

	// 写入显示ID (4字节)
	binary.BigEndian.PutUint32(buffer[24:28], 0)

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
	glog.Info(ctx, "视频设置已发送", "比特率:", settings.Bitrate, "最大帧率:", settings.MaxFps, "分辨率:", settings.Width, "x", settings.Height)
}

// sendTouchEvent 发送触摸事件
func (c *ScrcpyController) sendTouchEvent(ctx context.Context, tcpConn net.Conn, deviceConn *DeviceConnection, action, x, y int) {
	// 创建触摸事件消息
	buffer := make([]byte, 30) // 固定大小为30字节

	// 设置消息类型
	buffer[0] = TYPE_INJECT_TOUCH_EVENT

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
	if action == ACTION_DOWN {
		pressure = 0xFFFF
	}
	binary.BigEndian.PutUint16(buffer[22:24], pressure)

	// 写入按钮值 (4字节)
	binary.BigEndian.PutUint32(buffer[24:28], BUTTON_PRIMARY)

	// 发送消息
	_, err := tcpConn.Write(buffer)
	if err != nil {
		glog.Error(ctx, "发送触摸事件失败:", err)
		return
	}

	actionName := "未知"
	switch action {
	case ACTION_DOWN:
		actionName = "DOWN"
	case ACTION_UP:
		actionName = "UP"
	case ACTION_MOVE:
		actionName = "MOVE"
	}

	glog.Info(ctx, "触摸事件已发送",
		"类型:", actionName,
		"坐标:", x, y,
		"视频尺寸:", deviceConn.VideoWidth, "x", deviceConn.VideoHeight)
}

// sendSwipeEvent 发送滑动事件
func (c *ScrcpyController) sendSwipeEvent(ctx context.Context, tcpConn net.Conn, deviceConn *DeviceConnection, event SwipeEvent) {
	glog.Info(ctx, "开始滑动:", "起点:", event.StartX, event.StartY, "终点:", event.EndX, event.EndY)

	// 发送按下事件
	c.sendTouchEvent(ctx, tcpConn, deviceConn, ACTION_DOWN, event.StartX, event.StartY)
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
		c.sendTouchEvent(ctx, tcpConn, deviceConn, ACTION_MOVE, currentX, currentY)
		time.Sleep(time.Duration(stepDelay) * time.Millisecond)
	}

	// 发送抬起事件
	c.sendTouchEvent(ctx, tcpConn, deviceConn, ACTION_UP, event.EndX, event.EndY)
	glog.Info(ctx, "滑动事件已完成")
}

// checkPortForward 检查设备的端口转发是否正确设置
func (c *ScrcpyController) checkPortForward(ctx context.Context, udid string, port int) error {
	cmd := exec.Command("adb", "forward", "--list")
	output, err := cmd.Output()
	if err != nil {
		glog.Error(ctx, "[DEBUG] 无法获取ADB端口转发列表:", err)
		return fmt.Errorf("无法获取ADB端口转发列表: %w", err)
	}

	// 检查是否存在指定设备和端口的转发
	glog.Info(ctx, "[DEBUG] 当前ADB端口转发列表:")
	lines := strings.Split(string(output), "\n")
	foundForward := false

	// 支持多种设备ID格式:
	// 1. 完全匹配 "172.17.1.205"
	// 2. 包含端口的格式 "172.17.1.205:5555"
	deviceIdPattern := udid
	if !strings.Contains(udid, ":") {
		deviceIdPattern = udid + ":"
	}

	expectedPort := fmt.Sprintf("tcp:%d", port)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			glog.Info(ctx, "[DEBUG]", line)

			// 先检查设备ID（允许带端口号的格式）
			if strings.HasPrefix(line, udid) || strings.HasPrefix(line, deviceIdPattern) {
				// 再检查是否包含正确的端口转发
				if strings.Contains(line, expectedPort) {
					foundForward = true
					glog.Info(ctx, "[DEBUG] 已找到设备", udid, "的端口转发:", line)
					break
				}
			}
		}
	}

	if !foundForward {
		glog.Error(ctx, "[DEBUG] 没有找到设备", udid, "到端口", port, "的转发")
		return fmt.Errorf("设备 %s 没有设置到端口 %d 的转发", udid, port)
	}

	return nil
}

// connectToDevice 连接到设备的ADB转发端口 - 修改函数，从中移除端口转发检查部分
func (c *ScrcpyController) connectToDevice(ctx context.Context, udid string, port int) (net.Conn, error) {
	// 检查是否已经有连接
	if conn, ok := c.deviceConnections.Load(udid); ok {
		deviceConn := conn.(*DeviceConnection)
		// 如果端口相同且连接仍然有效，复用连接
		if deviceConn.Port == port {
			// 简单测试连接是否仍然有效
			err := deviceConn.Conn.SetDeadline(time.Now().Add(time.Second))
			if err == nil {
				err = deviceConn.Conn.SetDeadline(time.Time{}) // 重置deadline
				if err == nil {
					glog.Info(ctx, "[DEBUG] 复用已有WebSocket连接")
					return deviceConn.Conn, nil
				}
			}
			// 连接已失效，关闭并创建新连接
			deviceConn.Conn.Close()
		}
	}

	// 创建到设备的WebSocket连接
	deviceWSURL := fmt.Sprintf("ws://localhost:%d", port)
	glog.Info(ctx, "[DEBUG] 尝试连接到设备WebSocket:", deviceWSURL)

	// 创建WebSocket拨号器
	dialer := websocket.Dialer{
		HandshakeTimeout: 3 * time.Second,
	}

	// 连接到设备端WebSocket
	deviceWSConn, _, err := dialer.Dial(deviceWSURL, nil)
	if err != nil {
		glog.Error(ctx, "[DEBUG] WebSocket连接失败:", err)
		return nil, fmt.Errorf("连接到设备WebSocket失败: %w", err)
	}

	// 创建WebSocket连接的适配器，使其实现net.Conn接口
	wsAdapter := &WebSocketAdapter{
		conn: deviceWSConn,
	}

	glog.Info(ctx, "已连接到设备WebSocket", "设备ID:", udid, "端口:", port)
	return wsAdapter, nil
}

// WebSocketAdapter 实现net.Conn接口，适配WebSocket连接
type WebSocketAdapter struct {
	conn       *websocket.Conn
	readBuffer []byte
	readMutex  sync.Mutex
	writeMutex sync.Mutex
}

// Read 实现io.Reader接口
func (a *WebSocketAdapter) Read(b []byte) (n int, err error) {
	a.readMutex.Lock()
	defer a.readMutex.Unlock()

	// 如果缓冲区中还有数据，从缓冲区读取
	if len(a.readBuffer) > 0 {
		n = copy(b, a.readBuffer)
		a.readBuffer = a.readBuffer[n:]
		return n, nil
	}

	// 从WebSocket读取新消息
	messageType, p, err := a.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	// 只处理二进制消息
	if messageType != websocket.BinaryMessage {
		return 0, fmt.Errorf("非二进制消息: %d", messageType)
	}

	// 复制消息内容到目标缓冲区
	n = copy(b, p)

	// 如果消息长度超过目标缓冲区大小，保存剩余部分
	if n < len(p) {
		a.readBuffer = p[n:]
	}

	return n, nil
}

// Write 实现io.Writer接口
func (a *WebSocketAdapter) Write(b []byte) (n int, err error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	// 发送二进制消息
	err = a.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close 实现io.Closer接口
func (a *WebSocketAdapter) Close() error {
	return a.conn.Close()
}

// LocalAddr 实现net.Conn接口
func (a *WebSocketAdapter) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

// RemoteAddr 实现net.Conn接口
func (a *WebSocketAdapter) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

// SetDeadline 实现net.Conn接口
func (a *WebSocketAdapter) SetDeadline(t time.Time) error {
	err := a.conn.SetReadDeadline(t)
	if err != nil {
		return err
	}
	return a.conn.SetWriteDeadline(t)
}

// SetReadDeadline 实现net.Conn接口
func (a *WebSocketAdapter) SetReadDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}

// SetWriteDeadline 实现net.Conn接口
func (a *WebSocketAdapter) SetWriteDeadline(t time.Time) error {
	return a.conn.SetWriteDeadline(t)
}

// CleanupConnections 清理过期的连接
func (c *ScrcpyController) CleanupConnections() {
	// 定期清理过期的连接
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	now := time.Now()
	c.deviceConnections.Range(func(key, value interface{}) bool {
		deviceConn := value.(*DeviceConnection)
		// 如果连接超过30分钟未使用，关闭它
		if now.Sub(deviceConn.LastUsed) > 30*time.Minute {
			deviceConn.Conn.Close()
			c.deviceConnections.Delete(key)
			g.Log().Info(context.Background(), "已清理过期连接", "设备ID:", deviceConn.UdId)
		}
		return true
	})
}

// 用于解析 Exp-Golomb 编码的工具函数
func readUe(data []byte, nLen uint, startBit *uint) uint {
	// 计算前导零的个数
	nZeroNum := 0
	for *startBit < nLen*8 {
		if (data[*startBit/8] & (0x80 >> (*startBit % 8))) > 0 {
			break
		}
		nZeroNum++
		*startBit++
	}
	*startBit++

	// 计算结果
	dwRet := uint(0)
	for i := 0; i < nZeroNum; i++ {
		dwRet <<= 1
		if (data[*startBit/8] & (0x80 >> (*startBit % 8))) > 0 {
			dwRet += 1
		}
		*startBit++
	}
	return (1 << uint(nZeroNum)) - 1 + dwRet
}

// 读取有符号的 Exp-Golomb 编码
func readSe(data []byte, nLen uint, startBit *uint) int {
	ueVal := readUe(data, nLen, startBit)
	k := int64(ueVal)
	nValue := int(math.Ceil(float64(k) / 2))
	if ueVal%2 == 0 {
		nValue = -nValue
	}
	return nValue
}

// 读取指定位数的比特
func readBits(bitCount uint, data []byte, startBit *uint) uint {
	dwRet := uint(0)
	for i := uint(0); i < bitCount; i++ {
		dwRet <<= 1
		if (data[*startBit/8] & (0x80 >> (*startBit % 8))) > 0 {
			dwRet += 1
		}
		*startBit++
	}
	return dwRet
}

// 处理防竞争机制
func deEmulationPrevention(data []byte) []byte {
	size := len(data)
	tmpData := make([]byte, size)
	copy(tmpData, data)

	j := 0
	for i := 0; i < size-2; i++ {
		if i+2 < size &&
			tmpData[i] == 0x00 &&
			tmpData[i+1] == 0x00 &&
			tmpData[i+2] == 0x03 {
			tmpData[j] = tmpData[i]
			tmpData[j+1] = tmpData[i+1]
			i += 2
			j += 2
		} else {
			tmpData[j] = tmpData[i]
			j++
		}
	}
	// 复制剩余的字节
	for i := size - 2; i < size; i++ {
		if j < size {
			tmpData[j] = data[i]
			j++
		}
	}
	return tmpData[:j]
}

type SPSInfo struct {
	ProfileIdc                uint
	ConstraintSetFlags        uint
	LevelIdc                  uint
	PicWidthInMbsMinus1       uint
	PicHeightInMapUnitsMinus1 uint
	FrameMbsOnlyFlag          uint
	FrameCropLeftOffset       uint
	FrameCropRightOffset      uint
	FrameCropTopOffset        uint
	FrameCropBottomOffset     uint
	Sar                       [2]uint // Sample Aspect Ratio
}

func parseSPS(data []byte) (*SPSInfo, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("invalid SPS data length")
	}

	// 去除防竞争字节
	cleanData := deEmulationPrevention(data[4:]) // 跳过 NAL header
	var startBit uint = 0
	nLen := uint(len(cleanData))

	info := &SPSInfo{}

	// 跳过 forbidden_zero_bit, nal_ref_idc
	readBits(3, cleanData, &startBit)

	// 检查 NAL unit type 是否为 SPS (7)
	nalUnitType := readBits(5, cleanData, &startBit)
	if nalUnitType != 7 {
		return nil, fmt.Errorf("not a SPS NAL unit")
	}

	info.ProfileIdc = readBits(8, cleanData, &startBit)
	info.ConstraintSetFlags = readBits(8, cleanData, &startBit)
	info.LevelIdc = readBits(8, cleanData, &startBit)

	// 跳过 seq_parameter_set_id
	readUe(cleanData, nLen, &startBit)

	// 处理高级配置
	if info.ProfileIdc == 100 || info.ProfileIdc == 110 ||
		info.ProfileIdc == 122 || info.ProfileIdc == 144 {
		chromaFormatIdc := readUe(cleanData, nLen, &startBit)
		if chromaFormatIdc == 3 {
			readBits(1, cleanData, &startBit) // residual_colour_transform_flag
		}
		readUe(cleanData, nLen, &startBit) // bit_depth_luma_minus8
		readUe(cleanData, nLen, &startBit) // bit_depth_chroma_minus8
		readBits(1, cleanData, &startBit)  // qpprime_y_zero_transform_bypass_flag

		seqScalingMatrixPresentFlag := readBits(1, cleanData, &startBit)
		if seqScalingMatrixPresentFlag > 0 {
			for i := 0; i < 8; i++ {
				if readBits(1, cleanData, &startBit) > 0 {
					// 跳过缩放列表
					for j := 0; j < 64; j++ {
						readUe(cleanData, nLen, &startBit)
					}
				}
			}
		}
	}

	readUe(cleanData, nLen, &startBit) // log2_max_frame_num_minus4
	picOrderCntType := readUe(cleanData, nLen, &startBit)

	if picOrderCntType == 0 {
		readUe(cleanData, nLen, &startBit)
	} else if picOrderCntType == 1 {
		readBits(1, cleanData, &startBit)
		readSe(cleanData, nLen, &startBit)
		readSe(cleanData, nLen, &startBit)
		numRefFramesInPicOrderCntCycle := readUe(cleanData, nLen, &startBit)
		for i := uint(0); i < numRefFramesInPicOrderCntCycle; i++ {
			readSe(cleanData, nLen, &startBit)
		}
	}

	readUe(cleanData, nLen, &startBit) // max_num_ref_frames
	readBits(1, cleanData, &startBit)  // gaps_in_frame_num_value_allowed_flag

	info.PicWidthInMbsMinus1 = readUe(cleanData, nLen, &startBit)
	info.PicHeightInMapUnitsMinus1 = readUe(cleanData, nLen, &startBit)

	info.FrameMbsOnlyFlag = readBits(1, cleanData, &startBit)
	if info.FrameMbsOnlyFlag == 0 {
		readBits(1, cleanData, &startBit) // mb_adaptive_frame_field_flag
	}

	readBits(1, cleanData, &startBit) // direct_8x8_inference_flag

	frameCroppingFlag := readBits(1, cleanData, &startBit)
	if frameCroppingFlag > 0 {
		info.FrameCropLeftOffset = readUe(cleanData, nLen, &startBit)
		info.FrameCropRightOffset = readUe(cleanData, nLen, &startBit)
		info.FrameCropTopOffset = readUe(cleanData, nLen, &startBit)
		info.FrameCropBottomOffset = readUe(cleanData, nLen, &startBit)
	}

	vuiParametersPresentFlag := readBits(1, cleanData, &startBit)
	if vuiParametersPresentFlag > 0 {
		aspectRatioInfoPresentFlag := readBits(1, cleanData, &startBit)
		if aspectRatioInfoPresentFlag > 0 {
			aspectRatioIdc := readBits(8, cleanData, &startBit)
			if aspectRatioIdc == 255 { // Extended_SAR
				info.Sar[0] = readBits(16, cleanData, &startBit) // sar_width
				info.Sar[1] = readBits(16, cleanData, &startBit) // sar_height
			} else if aspectRatioIdc < 17 {
				// 使用预定义的 SAR 值
				info.Sar = getPredefinedSAR(aspectRatioIdc)
			}
		}
	}

	return info, nil
}

// 获取预定义的 SAR 值
func getPredefinedSAR(aspectRatioIdc uint) [2]uint {
	// 根据 H.264 标准定义的预设 SAR 值
	predefinedSAR := [][2]uint{
		{0, 0},    // Unspecified
		{1, 1},    // 1:1
		{12, 11},  // 12:11
		{10, 11},  // 10:11
		{16, 11},  // 16:11
		{40, 33},  // 40:33
		{24, 11},  // 24:11
		{20, 11},  // 20:11
		{32, 11},  // 32:11
		{80, 33},  // 80:33
		{18, 11},  // 18:11
		{15, 11},  // 15:11
		{64, 33},  // 64:33
		{160, 99}, // 160:99
		{4, 3},    // 4:3
		{3, 2},    // 3:2
		{2, 1},    // 2:1
	}

	if aspectRatioIdc < uint(len(predefinedSAR)) {
		return predefinedSAR[aspectRatioIdc]
	}
	return [2]uint{0, 0}
}
