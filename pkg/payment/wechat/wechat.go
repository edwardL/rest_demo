package wechat

import (
	"context"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WeChatPayConfig struct {
	AppID          string
	MchID          string
	MchSerialNo    string
	APIKey         string
	NotifyURL      string
	PrivateKeyPath string
}

type JSAPIClient struct {
	*core.Client
	AppID     string
	MchID     string
	APIKey    string
	NotifyURL string
}

func NewJSAPIClient(config *WeChatPayConfig) *JSAPIClient {
	// 加载商户私钥
	// privateKey, err := utils.LoadPrivateKeyWithPath(config.PrivateKeyPath)
	// if err != nil {
	// 	log.Fatal("load merchant private key error:", err)
	// }

	// 加载微信支付平台证书（示例中简化处理，实际应从微信获取）
	// 这里应该是从微信支付平台下载的证书内容
	// wechatPayCert := ""
	// certificate, err := utils.LoadCertificate(wechatPayCert)
	// if err != nil {
	// 	log.Fatal("load wechatpay certificate error:", err)
	// }

	// ctx := context.Background()
	// opts := []core.ClientOption{
	// 	option.WithWechatPayAuthCipher(config.MchID, config.MchSerialNo, privateKey, []*x509.Certificate{certificate}),
	// 	option.WithWechatPayCertificate([]*x509.Certificate{certificate}),
	// }

	// client, err := core.NewClient(ctx)
	// if err != nil {
	// 	log.Fatal("new wechatpay client err:", err)
	// }

	return &JSAPIClient{
		Client:    nil,
		AppID:     config.AppID,
		MchID:     config.MchID,
		APIKey:    config.APIKey,
		NotifyURL: config.NotifyURL,
	}
}

func (c *JSAPIClient) CreateJSAPIPayment(orderID, description, openID string, amount int64) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	svc := jsapi.JsapiApiService{Client: c.Client}

	resp, _, err := svc.PrepayWithRequestPayment(context.Background(),
		jsapi.PrepayRequest{
			Appid:       core.String(c.AppID),
			Mchid:       core.String(c.MchID),
			Description: core.String(description),
			OutTradeNo:  core.String(orderID),
			NotifyUrl:   core.String(c.NotifyURL),
			Amount: &jsapi.Amount{
				Total: core.Int64(amount),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openID),
			},
		},
	)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaPrivateKey, nil
}

// GenerateJSAPIParams 生成前端调起支付所需的参数
func (c *JSAPIClient) GenerateJSAPIParams(prepayID string) (map[string]string, error) {
	nonceStr := generateNonceStr()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	packageStr := fmt.Sprintf("prepay_id=%s", prepayID)

	// 按照字段名的ASCII码从小到大排序
	params := map[string]string{
		"appId":     c.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "RSA",
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接签名字符串
	var signStr string
	for _, k := range keys {
		signStr += k + "=" + params[k] + "&"
	}
	signStr = strings.TrimRight(signStr, "&")

	// 使用商户私钥进行签名
	privateKey, _ := LoadPrivateKey(c.APIKey)
	signature, err := utils.SignSHA256WithRSA(signStr, privateKey)
	if err != nil {
		return nil, err
	}

	params["paySign"] = signature
	return params, nil
}

func generateNonceStr() string {
	rand.Seed(time.Now().UnixNano())
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%d%d", rand.Intn(10000), time.Now().UnixNano())))
	return hex.EncodeToString(hash.Sum(nil))[:32]
}

// func (c *JSAPIClient) HandlePaymentNotify(header http.Header, body []byte) (*notify.Resource, error) {
// 	handler := notify.NewNotifyHandler(c.APIKey, c.wechatPayCert)

// 	resource := new(notify.Resource)
// 	notifyReq, err := handler.ParseNotifyRequest(context.Background(), header, body, resource)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 验证通知的合法性
// 	if notifyReq.EventType != "TRANSACTION.SUCCESS" {
// 		return nil, fmt.Errorf("invalid event type")
// 	}

// 	// 这里可以添加业务逻辑处理
// 	// 比如更新订单状态等

// 	return resource, nil
// }
