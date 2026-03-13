package payermax

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type NotifyType string

const (
	NotifyTypePayment NotifyType = "PAYMENT"
	NotifyTypeRefund  NotifyType = "REFUND"
)

type NotifyCommon struct {
	Code       string          `json:"code"`
	Msg        string          `json:"msg"`
	KeyVersion string          `json:"keyVersion"`
	AppId      string          `json:"appId"`
	MerchantNo string          `json:"merchantNo"`
	NotifyTime time.Time       `json:"notifyTime"`
	NotifyType NotifyType      `json:"notifyType"`
	Data       json.RawMessage `json:"data"`
}

type NotifyPayment struct {
	OutTradeNo     string    `json:"outTradeNo"`
	TradeToken     string    `json:"tradeToken"`
	TotalAmount    float64   `json:"totalAmount"`
	Currency       string    `json:"currency"`
	Country        string    `json:"country"`
	Status         string    `json:"status"`
	CompleteTime   time.Time `json:"completeTime"`
	PaymentDetails []struct {
		PaymentMethodType string `json:"paymentMethodType"`
		TargetOrg         string `json:"targetOrg"`
	} `json:"paymentDetails"`
	Reference string `json:"reference"`
}

type NotifyRefund struct {
	OutRefundNo      string    `json:"outRefundNo"`
	RefundTradeNo    string    `json:"refundTradeNo"`
	OutTradeNo       string    `json:"outTradeNo"`
	RefundAmount     float64   `json:"refundAmount"`
	RefundCurrency   string    `json:"refundCurrency"`
	RefundFinishTime time.Time `json:"refundFinishTime"`
	Status           string    `json:"status"`
}

func (p *Payermax) VerifyNotify(req *http.Request) error {
	responseSign := req.Header.Get("sign")
	// Read the content
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
	} else {
		return errors.New("cannot verify notify for HTTP Request with empty body")
	}
	// Restore the io.ReadCloser to its original state
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return VerifySign(bodyBytes, responseSign, p.payermaxPublicKey)
}
