package scrcpy_ws

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"

	"backend/internal/model"
	"backend/utility/h264"
)

// ScrcpyService 处理scrcpy的WebSocket连接和消息转发
type ScrcpyService struct {
	// 设备ID到TCP连接的映射
	deviceConnections sync.Map
	// 保护deviceConnections的互斥锁
	connMutex sync.RWMutex
}

// NewScrcpyService 创建一个新的ScrcpyService
func NewScrcpyService() *ScrcpyService {
	return &ScrcpyService{}
}

// HandleConnection 处理WebSocket连接
func (s *ScrcpyService) HandleConnection(ctx context.Context, wsConn *websocket.Conn, udid string, port int) error {
	glog.Info(ctx, "收到WebSocket连接请求", "设备ID:", udid, "端口:", port)

	// 预先检查ADB端口转发是否正确设置
	if err := s.checkPortForward(ctx, udid, port); err != nil {
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

		return err
	}

	// 连接到ADB转发的端口
	tcpConn, err := s.connectToDevice(ctx, udid, port)
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

		return err
	}
	defer tcpConn.Close()

	// 记录连接信息
	deviceConn := &model.DeviceConnection{
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
	s.deviceConnections.Store(udid, deviceConn)
	defer s.deviceConnections.Delete(udid)

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
	var doneClosed bool
	var doneMutex sync.Mutex

	// 从WebSocket接收消息并转发到TCP连接
	go func() {
		defer func() {
			// 使用recover防止关闭已关闭通道导致的panic
			if r := recover(); r != nil {
				glog.Warning(ctx, "关闭通道时发生panic:", r)
			}

			// 安全地关闭通道，确保只关闭一次
			doneMutex.Lock()
			if !doneClosed {
				close(done)
				doneClosed = true
			}
			doneMutex.Unlock()
		}()

		for {
			// 读取WebSocket消息
			messageType, message, err := wsConn.ReadMessage()
			if err != nil {
				// 检查是否是正常的连接关闭
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					glog.Debug(ctx, "前端WebSocket连接正常关闭")
				} else {
					glog.Error(ctx, "读取前端WebSocket消息失败:", err)
				}
				return
			}

			// 更新最后使用时间
			deviceConn.LastUsed = time.Now()

			// 处理JSON命令消息
			if messageType == websocket.TextMessage {
				s.handleCommandMessage(ctx, wsConn, tcpConn, deviceConn, message)
				continue
			}

			// 处理二进制消息
			if messageType == websocket.BinaryMessage {
				// 直接转发到TCP连接
				_, err = tcpConn.Write(message)
				if err != nil {
					// 检查是否是连接已关闭的错误
					if strings.Contains(err.Error(), "use of closed network connection") {
						glog.Debug(ctx, "与设备的连接已关闭")
					} else {
						glog.Error(ctx, "向设备连接写入数据失败:", err)
					}
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
				// 检查是否是连接已关闭的错误
				if strings.Contains(err.Error(), "use of closed network connection") {
					glog.Debug(ctx, "与设备的连接已正常关闭")
				} else {
					glog.Error(ctx, "从设备连接读取数据失败:", err)
				}

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
					// 安全地关闭通道，确保只关闭一次
					doneMutex.Lock()
					if !doneClosed {
						close(done)
						doneClosed = true
					}
					doneMutex.Unlock()
					return
				}
			}

			// 更新最后使用时间
			deviceConn.LastUsed = time.Now()

			// 获取接收到的数据
			received := buffer[:n]

			// 处理特殊消息类型
			if s.handleSpecialMessages(ctx, wsConn, deviceConn, received) {
				continue
			}

			// 转发到WebSocket
			err = wsConn.WriteMessage(websocket.BinaryMessage, received)
			if err != nil {
				glog.Error(ctx, "向WebSocket写入数据失败:", err)
				return
			}
		}
	}()

	// 等待直到连接关闭
	<-done
	glog.Info(ctx, "WebSocket连接已关闭", "设备ID:", udid)
	return nil
}

// handleCommandMessage 处理从客户端发来的JSON命令消息
func (s *ScrcpyService) handleCommandMessage(ctx context.Context, wsConn *websocket.Conn, tcpConn net.Conn, deviceConn *model.DeviceConnection, message []byte) {
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
		var touchEvent model.TouchEvent
		if data, ok := command["data"].(map[string]interface{}); ok {
			// 检查各字段是否存在
			xVal, xOk := data["x"].(float64)
			yVal, yOk := data["y"].(float64)
			actionVal, actionOk := data["action"].(float64)

			if !xOk || !yOk || !actionOk {
				glog.Error(ctx, "触摸事件缺少必要字段或字段类型错误",
					"x存在:", xOk,
					"y存在:", yOk,
					"action存在:", actionOk)
				return
			}

			touchEvent.Action = int(actionVal)
			touchEvent.X = int(xVal)
			touchEvent.Y = int(yVal)
			s.sendTouchEvent(ctx, tcpConn, deviceConn, touchEvent.Action, touchEvent.X, touchEvent.Y)
		}
	case "click":
		// 处理点击事件
		var clickEvent model.ClickEvent
		if data, ok := command["data"].(map[string]interface{}); ok {
			// 检查各字段是否存在
			xVal, xOk := data["x"].(float64)
			yVal, yOk := data["y"].(float64)

			if !xOk || !yOk {
				glog.Error(ctx, "点击事件缺少必要字段或字段类型错误",
					"x存在:", xOk,
					"y存在:", yOk)
				return
			}

			clickEvent.X = int(xVal)
			clickEvent.Y = int(yVal)

			// duration是可选字段，有默认值
			if durationVal, durationOk := data["duration"].(float64); durationOk {
				clickEvent.Duration = int(durationVal)
			}

			s.sendClickEvent(ctx, tcpConn, deviceConn, clickEvent)
		}
	case "swipe":
		// 处理滑动事件
		var swipeEvent model.SwipeEvent
		if data, ok := command["data"].(map[string]interface{}); ok {
			// 检查各字段是否存在
			startXVal, startXOk := data["startX"].(float64)
			startYVal, startYOk := data["startY"].(float64)
			endXVal, endXOk := data["endX"].(float64)
			endYVal, endYOk := data["endY"].(float64)

			if !startXOk || !startYOk || !endXOk || !endYOk {
				glog.Error(ctx, "滑动事件缺少必要字段或字段类型错误",
					"startX存在:", startXOk,
					"startY存在:", startYOk,
					"endX存在:", endXOk,
					"endY存在:", endYOk)
				return
			}

			swipeEvent.StartX = int(startXVal)
			swipeEvent.StartY = int(startYVal)
			swipeEvent.EndX = int(endXVal)
			swipeEvent.EndY = int(endYVal)

			// duration和steps是可选字段，有默认值
			if durationVal, durationOk := data["duration"].(float64); durationOk {
				swipeEvent.Duration = int(durationVal)
			}

			if stepsVal, stepsOk := data["steps"].(float64); stepsOk {
				swipeEvent.Steps = int(stepsVal)
			}

			s.sendSwipeEvent(ctx, tcpConn, deviceConn, swipeEvent)
		}
	case "videoSettings":
		// 处理视频设置
		var videoSettings model.VideoSettings
		if data, ok := command["data"].(map[string]interface{}); ok {
			// 检查各字段是否存在
			bitrateVal, bitrateOk := data["bitrate"].(float64)
			maxFpsVal, maxFpsOk := data["maxFps"].(float64)
			iFrameIntervalVal, iFrameIntervalOk := data["iFrameInterval"].(float64)
			widthVal, widthOk := data["width"].(float64)
			heightVal, heightOk := data["height"].(float64)

			if !bitrateOk || !maxFpsOk || !iFrameIntervalOk || !widthOk || !heightOk {
				glog.Error(ctx, "视频设置缺少必要字段或字段类型错误",
					"bitrate存在:", bitrateOk,
					"maxFps存在:", maxFpsOk,
					"iFrameInterval存在:", iFrameIntervalOk,
					"width存在:", widthOk,
					"height存在:", heightOk)
				return
			}

			videoSettings.Bitrate = int(bitrateVal)
			videoSettings.MaxFps = int(maxFpsVal)
			videoSettings.IFrameInterval = int(iFrameIntervalVal)
			videoSettings.Width = int(widthVal)
			videoSettings.Height = int(heightVal)
			s.sendVideoSettings(ctx, tcpConn, deviceConn, videoSettings)
		}
	default:
		glog.Warning(ctx, "未知命令类型:", cmdType)
	}
}

// handleSpecialMessages 处理特殊消息类型
func (s *ScrcpyService) handleSpecialMessages(ctx context.Context, wsConn *websocket.Conn, deviceConn *model.DeviceConnection, data []byte) bool {
	// 标记是否处理了特殊消息
	isSpecialMessage := false

	// 检查是否是初始化消息
	if len(data) > len(model.MAGIC_BYTES_INITIAL) && strings.HasPrefix(string(data), model.MAGIC_BYTES_INITIAL) {
		s.handleInitialInfo(ctx, wsConn, deviceConn, data)
		isSpecialMessage = true
		// 注意：handleInitialInfo已经负责转发消息，不需要在这里再转发
		return isSpecialMessage
	}

	// 检查是否是设备消息
	if len(data) > len(model.MAGIC_BYTES_MESSAGE) && strings.HasPrefix(string(data), model.MAGIC_BYTES_MESSAGE) {
		s.handleDeviceMessage(ctx, wsConn, deviceConn, data)
		isSpecialMessage = true
		// 注意：handleDeviceMessage已经负责转发消息，不需要在这里再转发
		return isSpecialMessage
	}

	// 检查是否是 SPS 数据
	if len(data) >= 5 {
		nalType := data[4] & 0x1F
		if nalType == 7 { // NAL type 7 表示 SPS
			spsInfo, err := h264.ParseSPS(data)
			if err != nil {
				glog.Error(ctx, "解析 SPS 失败", "error", err)
				// 返回false，让后续代码进行转发
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

			// 重要变更：不返回true，以确保SPS帧也能被转发
			isSpecialMessage = false
		}

		// // 也可以添加对PPS帧的识别和处理，但不拦截
		// if nalType == 8 { // NAL type 8 表示 PPS
		// 	glog.Debug(ctx, "检测到PPS帧，确保转发")
		// 	// 不做特殊处理，只是记录日志
		// 	isSpecialMessage = false
		// }
	}

	// 返回false，表示即使处理了特殊消息，也应该继续原始转发
	return isSpecialMessage
}

// handleInitialInfo 处理初始化信息
func (s *ScrcpyService) handleInitialInfo(ctx context.Context, wsConn *websocket.Conn, deviceConn *model.DeviceConnection, data []byte) {
	glog.Info(ctx, "处理初始化信息...")

	// 直接转发初始化信息到客户端
	err := wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		glog.Error(ctx, "转发初始化信息失败:", err)
		return
	}

	// 尝试解析屏幕尺寸和客户端ID，实际应用中可能需要更复杂的解析逻辑
	offset := len(model.MAGIC_BYTES_INITIAL)

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
		s.sendVideoSettings(ctx, deviceConn.Conn, deviceConn, model.VideoSettings{
			Bitrate:        model.BITRATE,
			MaxFps:         model.MAX_FPS,
			IFrameInterval: model.I_FRAME_INTERVAL,
			Width:          model.WIDTH,
			Height:         model.HEIGHT,
		})
	}
}

// handleDeviceMessage 处理设备消息
func (s *ScrcpyService) handleDeviceMessage(ctx context.Context, wsConn *websocket.Conn, deviceConn *model.DeviceConnection, data []byte) {
	glog.Info(ctx, "处理设备消息...")

	// 直接转发设备消息到客户端
	err := wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		glog.Error(ctx, "转发设备消息失败:", err)
		return
	}

	// 解析消息类型，后续可以添加特定消息类型的处理
	if len(data) > len(model.MAGIC_BYTES_MESSAGE) {
		msgType := data[len(model.MAGIC_BYTES_MESSAGE)]
		glog.Debug(ctx, "设备消息类型:", msgType)
	}
}

// checkPortForward 检查设备的端口转发是否正确设置
func (s *ScrcpyService) checkPortForward(ctx context.Context, udid string, port int) error {
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

// connectToDevice 连接到设备的ADB转发端口
func (s *ScrcpyService) connectToDevice(ctx context.Context, udid string, port int) (net.Conn, error) {
	// 检查是否已经有连接
	if conn, ok := s.deviceConnections.Load(udid); ok {
		deviceConn := conn.(*model.DeviceConnection)
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
func (s *ScrcpyService) CleanupConnections() {
	// 定期清理过期的连接
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	now := time.Now()
	s.deviceConnections.Range(func(key, value interface{}) bool {
		deviceConn := value.(*model.DeviceConnection)
		// 如果连接超过30分钟未使用，关闭它
		if now.Sub(deviceConn.LastUsed) > 30*time.Minute {
			deviceConn.Conn.Close()
			s.deviceConnections.Delete(key)
			g.Log().Info(context.Background(), "已清理过期连接", "设备ID:", deviceConn.UdId)
		}
		return true
	})
}
