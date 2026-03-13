package paypal

import (
	"encoding/json"
	"time"
)

const CodeSuccess = "SUCCESS"

// The order was created with the specified context.
const PaymentStatusCreated = "CREATED"

// The payment was authorized or the authorized payment was captured for the order.
const PaymentStatusCompleted = "COMPLETED"

const PaymentStatusRefunded = "REFUNDED"

// The customer approved the payment through the PayPal wallet or another form of guest or unbranded payment. For example, a card, bank account, or so on.
const PaymentStatusApproved = "APPROVED"

// All purchase units in the order are voided.
const PaymentStatusVoided = "VOIDED" //All purchase units in the order are voided.

// EventType https://developer.paypal.com/api/rest/webhooks/event-names/#orders-v2
type EventType string

const (
	//	A payment capture completes.
	EventTypePaymentCaptureCompleted EventType = "PAYMENT.CAPTURE.COMPLETED"
	//A buyer approved a checkout order
	EventTypeCheckoutOrderApproved EventType = "CHECKOUT.ORDER.APPROVED"
	//The state of a payment capture changes to pending.
	EventTypePaymentCapturePending EventType = "PAYMENT.CAPTURE.PENDING"
	// payment capture is denied.
	EventTypePaymentCaptureDenied EventType = "PAYMENT.CAPTURE.DENIED"
	//A merchant refunds a payment capture.
	EventTypePaymentCaptureRefunded EventType = "PAYMENT.CAPTURE.REFUNDED"
	//PayPal reverses a payment capture.
	EventTypePaymentCaptureReversed EventType = "PAYMENT.CAPTURE.REVERSED"
)

type NotifyCommon struct {
	Id              string          `json:"id"`
	EventVersion    string          `json:"event_version"`
	CreateTime      time.Time       `json:"create_time"`
	ResourceType    string          `json:"resource_type"`
	ResourceVersion string          `json:"resource_version"`
	EventType       EventType       `json:"event_type"`
	Summary         string          `json:"summary"`
	Resource        json.RawMessage `json:"resource"`
	Links           []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

// PAYMENT.CAPTURE.COMPLETED
type NotifyPaymentCaptureCompleted struct {
	Payee struct {
		EmailAddress string `json:"email_address"`
		MerchantId   string `json:"merchant_id"`
	} `json:"payee"`
	Amount struct {
		Value        string `json:"value"`
		CurrencyCode string `json:"currency_code"`
	} `json:"amount"`
	SellerProtection struct {
		DisputeCategories []string `json:"dispute_categories"`
		Status            string   `json:"status"`
	} `json:"seller_protection"`
	SupplementaryData struct {
		RelatedIds struct {
			OrderId string `json:"order_id"`
		} `json:"related_ids"`
	} `json:"supplementary_data"`
	UpdateTime                time.Time `json:"update_time"`
	CreateTime                time.Time `json:"create_time"`
	FinalCapture              bool      `json:"final_capture"`
	SellerReceivableBreakdown struct {
		PaypalFee struct {
			Value        string `json:"value"`
			CurrencyCode string `json:"currency_code"`
		} `json:"paypal_fee"`
		GrossAmount struct {
			Value        string `json:"value"`
			CurrencyCode string `json:"currency_code"`
		} `json:"gross_amount"`
		NetAmount struct {
			Value        string `json:"value"`
			CurrencyCode string `json:"currency_code"`
		} `json:"net_amount"`
	} `json:"seller_receivable_breakdown"`
	CustomId string `json:"custom_id"`
	Links    []struct {
		Method string `json:"method"`
		Rel    string `json:"rel"`
		Href   string `json:"href"`
	} `json:"links"`
	Id     string `json:"id"`
	Status string `json:"status"`
}

// CHECKOUT.ORDER.APPROVED
type NotifyCheckoutOrderApproved struct {
	CreateTime    time.Time `json:"create_time"`
	PurchaseUnits []struct {
		ReferenceId string `json:"reference_id"`
		Amount      struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"amount"`
		Payee struct {
			EmailAddress string `json:"email_address"`
			MerchantId   string `json:"merchant_id"`
			DisplayData  struct {
				BrandName string `json:"brand_name"`
			} `json:"display_data"`
		} `json:"payee"`
		CustomId string `json:"custom_id"`
		Shipping struct {
			Name struct {
				FullName string `json:"full_name"`
			} `json:"name"`
			Address struct {
				AddressLine1 string `json:"address_line_1"`
				AdminArea2   string `json:"admin_area_2"`
				AdminArea1   string `json:"admin_area_1"`
				PostalCode   string `json:"postal_code"`
				CountryCode  string `json:"country_code"`
			} `json:"address"`
		} `json:"shipping"`
	} `json:"purchase_units"`
	Links []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
	Id            string `json:"id"`
	PaymentSource struct {
		Paypal struct {
			EmailAddress  string `json:"email_address"`
			AccountId     string `json:"account_id"`
			AccountStatus string `json:"account_status"`
			Name          struct {
				GivenName string `json:"given_name"`
				Surname   string `json:"surname"`
			} `json:"name"`
			Address struct {
				CountryCode string `json:"country_code"`
			} `json:"address"`
		} `json:"paypal"`
	} `json:"payment_source"`
	Intent string `json:"intent"`
	Payer  struct {
		Name struct {
			GivenName string `json:"given_name"`
			Surname   string `json:"surname"`
		} `json:"name"`
		EmailAddress string `json:"email_address"`
		PayerId      string `json:"payer_id"`
		Address      struct {
			CountryCode string `json:"country_code"`
		} `json:"address"`
	} `json:"payer"`
	Status string `json:"status"`
}

// PAYMENT.CAPTURE.REFUNDED
type NotifyCheckoutOrderRefunded struct {
	Id       string
	CustomId string `json:"custom_id"`
	Amount   struct {
		CurrencyCode string `json:"currency_code"`
		Value        string `json:"value"`
	} `json:"amount"`
	SellerPayableBreakdown struct {
		GrossAmount struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"gross_amount"`
		PaypalFee struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"paypal_fee"`
		NetAmount struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"net_amount"`
		TotalRefundedAmount struct {
			CurrencyCode string `json:"currency_code"`
			Value        string `json:"value"`
		} `json:"total_refunded_amount"`
	} `json:"seller_payable_breakdown"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	Links      []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}
