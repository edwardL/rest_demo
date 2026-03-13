package payermax

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeebo/errs"
)

var Err = errs.Class("payment.payermax")

type Config struct {
	Endpoint   string        `help:"接口地址" default:"https://pay-gate.payermax.com/aggregate-pay/api/gateway"`
	AppId      string        `help:"appid" default:""`
	MerchantNo string        `help:"商户id" default:""`
	PrivateKey string        `help:"商户签名私钥" default:""`
	PublicKey  string        `help:"payermax签名校验公钥" default:""`
	Timeout    time.Duration `help:"请求超时时间" default:"30s"`
	Version    string        `help:"接口版本" default:"1.1"`
	KeyVersion string        `help:"密钥版本" default:"1"`
}

type Payermax struct {
	conf               Config
	httpClient         *http.Client
	merchantPrivateKey *rsa.PrivateKey // 商户私钥
	payermaxPublicKey  *rsa.PublicKey  // payermax公钥钥
}

func NewPayermax(conf Config) (client *Payermax, err error) {

	if conf.Endpoint == "" || conf.AppId == "" || conf.MerchantNo == "" || conf.PrivateKey == "" || conf.PublicKey == "" {
		return nil, Err.New("payermax config error")
	}
	priKey, err := DecodePrivateKey(conf.PrivateKey)
	if err != nil {
		return nil, Err.Wrap(err)
	}

	pubKey, err := DecodePublicKey(conf.PublicKey)
	if err != nil {
		return nil, Err.Wrap(err)
	}
	conf = initConfig(conf)
	client = &Payermax{
		conf: conf,
	}
	client.merchantPrivateKey = priKey
	client.payermaxPublicKey = pubKey
	client.httpClient = &http.Client{
		Timeout: conf.Timeout,
	}
	return client, nil
}

func (p *Payermax) send(apiName string, reqData any, resp Response) (err error) {
	if resp == nil {
		return errors.New("resp is nil")
	}
	reqBody := p.getRequestBody(reqData)
	resultBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", p.getApiUrl(apiName), bytes.NewBuffer(resultBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	rsaSign, err := GetRsaSign(resultBytes, p.merchantPrivateKey)
	if err != nil {
		return
	}
	req.Header.Set("sign", rsaSign)
	response, err := p.httpClient.Do(req)
	if err != nil {
		return
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("response code %d, body:%s", response.StatusCode, string(responseBytes))
		return
	}
	responseSign := response.Header.Get("sign")
	if err = VerifySign(responseBytes, responseSign, p.payermaxPublicKey); err != nil {
		return
	}
	err = json.Unmarshal(responseBytes, &resp)
	if err != nil {
		return
	}
	err = resp.Error()
	return
}

func (p *Payermax) getRequestBody(data any) CommonReq {
	return CommonReq{
		KeyVersion:  p.conf.KeyVersion,
		MerchantNo:  p.conf.MerchantNo,
		RequestTime: time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"),
		Version:     p.conf.Version,
		AppId:       p.conf.AppId,
		Data:        data,
	}
}

func (p *Payermax) getApiUrl(apiName string) string {
	return p.conf.Endpoint + "/" + strings.TrimLeft(apiName, "/")
}

func initConfig(conf Config) Config {
	if conf.KeyVersion == "" {
		conf.KeyVersion = "1"
	}
	if conf.Version == "" {
		conf.Version = "1.1"
	}
	if conf.Timeout <= 0 || conf.Timeout > time.Minute*60 {
		conf.Timeout = time.Second * 30
	}
	conf.Endpoint = strings.TrimRight(conf.Endpoint, "/")
	return conf
}
