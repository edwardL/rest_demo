package handler

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	v1 "rest_demo/api/v1"
	"rest_demo/internal/service"
	"rest_demo/pkg/payment/wechat"
	"rest_demo/pkg/websocket"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	loginService service.LoginService
	client       *wechat.JSAPIClient
}

func NewLoginHandler(s service.LoginService, client *wechat.JSAPIClient) *LoginHandler {
	return &LoginHandler{
		loginService: s,
		client:       client,
	}
}

// Login 登陆
// @Summary 登陆
// @Description 登陆成功后返回token
// @Tags 登陆
// @param json body v1.LoginReq true "请求数据"
// @Accept json
// @Produce json
// @Response 200 {object} v1.Resp{result=v1.LoginRes}
// @Router /login [post]
func (h *LoginHandler) Login(ctx *gin.Context) {
	req := v1.LoginReq{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.Error(ctx, err)
		return
	}
	res, err := h.loginService.Login(ctx, &req)
	v1.Response(ctx, err, res)
}

func (h *LoginHandler) CreateOrder(c *gin.Context) {
	var req struct {
		OpenID      string `json:"openid"`
		Amount      int64  `json:"amount"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成订单号(实际项目中应该有自己的订单生成逻辑)
	orderID := fmt.Sprintf("ORDER%d", time.Now().UnixNano())

	// 调用微信支付接口
	resp, err := h.client.CreateJSAPIPayment(orderID, req.Description, req.OpenID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 生成前端调起支付所需的参数
	jsapiParams, err := h.client.GenerateJSAPIParams(*resp.PrepayId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"params":   jsapiParams,
	})
}

func (h *LoginHandler) GetApiList(engine *gin.Engine) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		routesInfo := engine.Routes()
		list := make([]v1.ApiItem, 0)
		for i := range routesInfo {
			// if !strings.HasPrefix(routesInfo[i].Path, "/api/") {
			// 	continue
			// }
			v := routesInfo[i].Method + ":" + routesInfo[i].Path
			key := md5.Sum([]byte(v))
			list = append(list, v1.ApiItem{
				Method: routesInfo[i].Method,
				Path:   routesInfo[i].Path,
				Value:  fmt.Sprintf("%x", key),
				Label:  v,
			})
		}
		v1.Response(ctx, nil, list)
	}
}

// Wsplay 通过 websocket 播放 mpegts 数据
func (h *LoginHandler) Wsplay(c *gin.Context) {
	websocket.WsManager.RegisterClient(c)
}

var (
	cmd     *exec.Cmd
	cmdLock sync.Mutex
)

func (h *LoginHandler) Stream(c *gin.Context) {
	rtspUrl := "html/1365070268951.mp4" // 替换为你的RTSP源地址

	cmdLock.Lock()
	defer cmdLock.Unlock()

	// 如果已有FFmpeg进程在运行，先关闭
	if cmd != nil {
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill existing FFmpeg process: %v", err)
		}
	}
	// 启动ffmpeg 转码
	cmd = exec.Command("ffmpeg",
		"-i", rtspUrl,
		"-c", "copy", // 使用原始编解码器
		"-f", "flv", // 输出FLV格式
		"-an",    // 不处理音频(可选)
		"pipe:1", // 输出到标准输出
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := cmd.Start(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// 设置响应头
	c.Header("Content-Type", "video/x-flv")
	c.Header("Connection", "keep-alive")
	c.Header("Cache-Control", "no-cache")

	// 流式传输
	buf := make([]byte, 4096)
	for {
		n, err := stdout.Read(buf)
		if err != nil {
			break
		}
		if _, err := c.Writer.Write(buf[:n]); err != nil {
			break
		}
		c.Writer.Flush()
	}
}
