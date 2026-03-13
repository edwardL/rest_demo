package payermax

import "time"

type PaymentDetail struct {
	PaymentMethodType string `json:"paymentMethodType"`
	TargetOrg         string `json:"targetOrg"`
}

type OrderAndPayReq struct {
	OutTradeNo       string        `json:"outTradeNo"`
	Integrate        string        `json:"integrate"` //使用PayerMax_Hosted_Payment_Page进行付款方支付信息收集及处理，需传入参数：Hosted_Checkout
	Subject          string        `json:"subject"`
	TotalAmount      float64       `json:"totalAmount"`
	Currency         string        `json:"currency"`
	Country          string        `json:"country"`
	UserId           string        `json:"userId"`
	FrontCallbackUrl string        `json:"frontCallbackUrl"` //商户指定的跳转URL，用户完成支付后会被跳转到该地址，以http/https开头或者商户应用的scheme地址
	NotifyUrl        string        `json:"notifyUrl"`        //回调地址
	PaymentDetail    PaymentDetail `json:"payment_detail,omitempty"`
}

type OrderAndPayRes struct {
	CommonRes
	Data struct {
		RedirectUrl string `json:"redirectUrl"`
		OutTradeNo  string `json:"outTradeNo"`
		TradeToken  string `json:"tradeToken"`
		Status      string `json:"status"`
	}
}

type OrderQueryReq struct {
	OutTradeNo string `json:"outTradeNo"`
}

type OrderQueryRes struct {
	CommonRes

	Data struct {
		Reference      string    `json:"reference"`
		Country        string    `json:"country"`
		TotalAmount    float64   `json:"totalAmount"`
		OutTradeNo     string    `json:"outTradeNo"`
		Currency       string    `json:"currency"`
		ChannelNo      string    `json:"channelNo"`
		ThirdChannelNo string    `json:"thirdChannelNo"`
		PaymentCode    string    `json:"paymentCode"`
		TradeToken     string    `json:"tradeToken"`
		CompleteTime   time.Time `json:"completeTime"`
		PaymentDetails []struct {
			TargetOrg string `json:"targetOrg"`
			CardInfo  struct {
				CardOrg            string `json:"cardOrg"`
				Country            string `json:"country"`
				CardIdentifierNo   string `json:"cardIdentifierNo"`
				CardIdentifierName string `json:"cardIdentifierName"`
			} `json:"cardInfo"`
			PayAmount         float64 `json:"payAmount"`
			ExchangeRate      string  `json:"exchangeRate"`
			PaymentMethod     string  `json:"paymentMethod"`
			PayCurrency       string  `json:"payCurrency"`
			PaymentMethodType string  `json:"paymentMethodType"`
		} `json:"paymentDetails"`
		Fees struct {
			MerFee struct {
				Url      string `json:"url"`
				Amount   string `json:"amount"`
				Currency string `json:"currency"`
			} `json:"merFee"`
		} `json:"fees"`
		Status    string `json:"status"`
		ResultMsg string `json:"resultMsg"`
	}
}

type OrderRefundReq struct {
	OutRefundNo     string  `json:"outRefundNo"`
	RefundAmount    float64 `json:"refundAmount"`
	RefundCurrency  string  `json:"refundCurrency"`
	OutTradeNo      string  `json:"outTradeNo"`
	Comments        string  `json:"comments"`
	RefundNotifyUrl string  `json:"refundNotifyUrl"`
}

type OrderRefundRes struct {
	CommonRes
	Data struct {
		OutRefundNo   string `json:"outRefundNo"`
		TradeOrderNo  string `json:"tradeOrderNo"`
		RefundTradeNo string `json:"refundTradeNo"`
		Status        string `json:"status"`
	} `json:"data"`
}

// /https://docs.payermax.com/api.html?docName=New%20Version&docVer=v1.0&docLang=cn#/paths/aggregate-pay-api-gateway-orderAndPay/post
func (p *Payermax) OrderAndPay(req *OrderAndPayReq) (res *OrderAndPayRes, err error) {
	res = &OrderAndPayRes{}
	err = p.send("/orderAndPay", req, res)
	return
}

// https://docs.payermax.com/api.html?docName=New%20Version&docVer=v1.0&docLang=cn#/paths/aggregate-pay-api-gateway-orderQuery/post
func (p *Payermax) OrderQuery(req *OrderQueryReq) (res *OrderQueryRes, err error) {
	res = &OrderQueryRes{}
	err = p.send("/orderQuery", req, res)
	return
}

// https://docs.payermax.com/api.html?docName=New%20Version&docVer=v1.0&docLang=cn#/paths/aggregate-pay-api-gateway-refund/post
func (p *Payermax) OrderRefund(req *OrderRefundReq) (res *OrderRefundRes, err error) {
	res = &OrderRefundRes{}
	err = p.send("/refund", req, res)
	return
}
