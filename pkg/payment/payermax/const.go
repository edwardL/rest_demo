package payermax

import (
	"errors"
	"fmt"
)

const CodeSuccess = "APPLY_SUCCESS"
const CodeFailed = "PAYMENT_FAILED"
const CodeClosed = "ORDER_CLOSED"

const PaymentStatusSuccess = "SUCCESS"
const PaymentStatusFailed = "FAILED"
const PaymentStatusClosed = "CLOSED"
const RefundStatusSuccess = "REFUND_SUCCESS"

const (
	Uat  = "https://pay-gate-uat.payermax.com/aggregate-pay/api/gateway"
	Prod = "https://pay-gate.payermax.com/aggregate-pay/api/gateway"
)

type CommonReq struct {
	KeyVersion  string `json:"keyVersion"`
	MerchantNo  string `json:"merchantNo"`
	RequestTime string `json:"requestTime"`
	Version     string `json:"version"`
	AppId       string `json:"appId"`
	Data        any    `json:"data"`
}

type CommonRes struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func (r *CommonRes) GetCode() string {
	return r.Code
}
func (r *CommonRes) GetMsg() string {
	return r.Msg
}

func (r *CommonRes) Error() error {
	if r == nil {
		return errors.New("response is nil")
	}
	if r.Code != CodeSuccess {
		return fmt.Errorf("[%s]%s", r.Code, r.Msg)
	}
	return nil
}

type Response interface {
	GetCode() string
	GetMsg() string
	Error() error
}
