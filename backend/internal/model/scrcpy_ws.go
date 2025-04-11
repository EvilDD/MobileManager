package model

import (
	"net"
	"time"
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

// ClickEvent 点击事件结构
type ClickEvent struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Duration int `json:"duration"` // 点击持续时间(毫秒)
}

// VideoSettings 视频设置结构
type VideoSettings struct {
	Bitrate        int `json:"bitrate"`
	MaxFps         int `json:"maxFps"`
	IFrameInterval int `json:"iFrameInterval"`
	Width          int `json:"width"`
	Height         int `json:"height"`
}

// 视频流初始参数
const (
	BITRATE          = 5000000
	MAX_FPS          = 24
	I_FRAME_INTERVAL = 5
	WIDTH            = 540
	HEIGHT           = 960
)
