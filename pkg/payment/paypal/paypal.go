package paypal

import (
	paypal2 "github.com/plutov/paypal/v4"
)

type Config struct {
	Endpoint  string `help:"paypal支付接口地址" default:"https://api.paypal.com"`
	ClientId  string `help:"paypal支付clientId" default:""`
	Secret    string `help:"paypal支付secret" default:""`
	WebhookId string `help:"webhook id" default:""`
}

type Paypal struct {
	conf Config
	*paypal2.Client
}

func NewPaypal(conf Config) (*Paypal, error) {
	if conf.Endpoint == "" || conf.ClientId == "" || conf.Secret == "" {
		conf.Endpoint = paypal2.APIBaseLive
	}
	c, err := paypal2.NewClient(conf.ClientId, conf.Secret, conf.Endpoint)
	if err != nil {
		return nil, err
	}
	return &Paypal{Client: c, conf: conf}, nil
}

func (p *Paypal) GetWebhookId() string {
	return p.conf.WebhookId
}

// https://developer.paypal.com/docs/api/orders/v2/#orders_create
//func (p *Paypal) CreateOrder(product *web.Product, payer *web.Payer, order *web.Order) (res *web.CreateOrderRes, err error) {
//	defer func() {
//		if err != nil {
//			err = Err.Wrap(err)
//		}
//	}()
//	purUnits := []paypal2.PurchaseUnitRequest{
//		{
//			CustomID: order.OrderId,
//			Amount: &paypal2.PurchaseUnitAmount{
//				Currency: product.Currency,
//				Value:    fmt.Sprintf("%.2f", float32(product.Price)/100),
//			},
//		},
//	}
//	_payer := &paypal2.CreateOrderPayer{}
//	appContext := &paypal2.ApplicationContext{
//		BrandName: "Hoby",
//		ReturnURL: order.ReturnUrl,
//		CancelURL: order.ReturnUrl,
//	}
//	_, err = p.GetAccessToken()
//	if err != nil {
//		return
//	}
//	var resp *paypal2.Order
//	resp, err = p.Client.CreateOrder("CAPTURE", purUnits, _payer, appContext)
//	if err != nil {
//		return
//	}
//	if resp == nil || resp.Status != "CREATED" {
//		err = fmt.Errorf("create order error")
//		return
//	}
//	res = &web.CreateOrderRes{
//		PayUrl: resp.Links[1].Href,
//	}
//	return
//}
