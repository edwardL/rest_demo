package httpreq

import (
	"context"
	"net/url"
	"time"
)

// AfterCallbackFunc 完成回调函数，处理响应结果
// httpStatus: HTTP响应状态码
// body: 响应body内容
// 返回处理后的body和错误
type AfterCallbackFunc func(httpStatus int, body []byte) ([]byte, error)

// RequestArgs HTTP请求参数
// 包含请求的URL、参数、头信息、响应体和错误列表
type RequestArgs struct {
	// ReqUrl 请求地址（完整 URL）
	ReqUrl string `json:"req_url"`
	// ReqParams 请求参数（可为 query/body 对象）
	ReqParams any `json:"req_params"`
	// ReqHeaders 请求头（建议使用 map[string]string 或 map[string][]string）
	ReqHeaders any `json:"req_headers"`
	// Method HTTP 请求方法，例如 GET/POST/PUT/DELETE
	Method string `json:"method"`
	// ContentType 请求的内容类型，例如 application/json
	ContentType string `json:"content_type"`
	// HttpStatusCode 响应状态码
	HttpStatusCode int `json:"http_status_code"`
	// RespBody 响应体原始字节内容
	RespBody []byte `json:"resp_body"`
	// ErrorList 请求处理过程中产生的错误集合
	ErrorList []error `json:"error_list"`
}

// HttpReq HTTP 请求器配置
// 通过链式 SetXxx 方法设置请求基础参数、超时与重试策略。
type HttpReq struct {
	// baseUrl 基础地址（如 https://api.example.com）
	baseUrl string
	// timeout 单次请求超时时间
	timeout time.Duration
	// tryAgainNum 失败后的重试次数（不含首次请求）
	tryAgainNum int
	// tryAgainTime 每次重试前的等待时间
	tryAgainTime time.Duration
	// reqAfterCallback 请求完成后的响应处理回调
	reqAfterCallback AfterCallbackFunc
}

// NewReq 创建并返回默认配置的 HttpReq。
// 默认超时时间为 120 秒。
func NewReq() *HttpReq {
	return &HttpReq{
		baseUrl: "",
		timeout: 120 * time.Second,
	}
}

// SetAfterCallback 设置请求完成后的回调函数。
// 可用于统一处理响应码、响应体转换或错误封装。
func (req *HttpReq) SetAfterCallback(f AfterCallbackFunc) *HttpReq {
	req.reqAfterCallback = f
	return req
}

// SetTimeout 设置单次请求超时时间。
func (req *HttpReq) SetTimeout(timeout time.Duration) *HttpReq {
	req.timeout = timeout
	return req
}

// SetBaseUrl 设置请求基础地址。
// 业务请求可在此基础上拼接具体路径。
func (req *HttpReq) SetBaseUrl(uri string) *HttpReq {
	req.baseUrl = uri
	return req
}

// SetTryAgainNum 设置失败后的重试次数。
func (req *HttpReq) SetTryAgainNum(num int) *HttpReq {
	req.tryAgainNum = num
	return req
}

// SetTryAgainTime 设置每次重试间隔时间。
func (req *HttpReq) SetTryAgainTime(t time.Duration) *HttpReq {
	req.tryAgainTime = t
	return req
}

// PostJson 发送POST JSON请求
// ctx: 上下文
// uri: 请求路径（相对于baseUrl）
// payload: 请求载荷（map类型，会序列化为JSON）
// headers: 请求头
// 返回响应body、请求参数和错误
func (req *HttpReq) PostJSON(ctx context.Context, uri string, payload any, headers map[string]string) (body []byte, reqArgs *RequestArgs, err error) {
	return nil, nil, err
}

func BuildUrl(baseUrl string, apiUrl string) (string, error) {
	var fullUrl string = apiUrl
	var err error
	if baseUrl != "" {
		fullUrl, err = url.JoinPath(baseUrl, apiUrl)
		if err != nil {
			fullUrl = apiUrl
		}
	}
	return fullUrl, nil
}
