package websocket

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"

	"backend/internal/service/scrcpy_ws"
)

// ScrcpyController 处理scrcpy的WebSocket连接和消息转发
type ScrcpyController struct {
	// 升级HTTP连接到WebSocket的upgrader
	upgrader websocket.Upgrader
	// scrcpy服务
	scrcpyService *scrcpy_ws.ScrcpyService
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
		scrcpyService: scrcpy_ws.NewScrcpyService(),
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

	// 使用service层处理连接
	err = c.scrcpyService.HandleConnection(ctx, wsConn, udid, port)
	if err != nil {
		glog.Error(ctx, "处理WebSocket连接失败:", err)
		return
	}
}
