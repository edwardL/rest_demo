package payermax_test

import (
	"fmt"
	"rest_demo/pkg/payment/payermax"
	"testing"
)

func getPayermax(t *testing.T) *payermax.Payermax {
	p, err := payermax.NewPayermax(payermax.Config{
		Endpoint:   "https://pay-gate-uat.payermax.com/aggregate-pay/api/gateway",
		AppId:      "e4f3caba3413461bb85d576582293418",
		MerchantNo: "SDP01010114337492",
		PrivateKey: `MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCFF+ys52uDYJaVBH9EbMih+zTvmVUpiHO1BD5CKQLa/4dTtaTGvmeqy2x6J/uH1G4C6eAc/g7nCwM5qpnrc1UJWjRZ+ayaDQQq3wQJlaw5crkRx9qMlGPMXMSg607EXh8CjCbMkCkIJNcUBVW/K/oAsiMmaHYK3yF2XwTX/2s9QbBNok9yiH9fIHTbmarT4zxltZGy38VZTLyV6fMguWq4Ha2m9eJf4OyltgIM7LDShLO2r3+0LbnMPZP6OTUbyqYx5YLGQdmq5DA7YzLfpWfMgj9tLVVEOS4tdWNymUQD9W/Gb7sGmkUz+EGR/5Fzk6WnIsIhMXmbIci4ieNz9hURAgMBAAECggEAERwPP/cMGjqLgSKv3bMCY9hwaLDUgt9YyJeADW9KP68DzZ4XTbliiFFYY6fKLR8A+XzpY7DBZ7BBvOMSoHMWJnqjKkHvj2pG89/xm+3S0xvNcNy5WsMkxTvTx0AYwyO6ZtBvmHKb48EgqSE6cbYMkJRV8nURX7ppidcTP1VFiv08X4mJg/cI7UMbVsWsWQoxCr2ZxOVVSrhZne+pgnNgFocmnekucFH4mnwUIxdQwjw5mrhZsqI2W7jcIkRA4DvXDTwJj15IukQrPN0KGw7lFcKx4mFO8VA/HOF3i90Z8rMvisuYTBKQX2SYCF6T1YPEmfZxOC115EuVo4+Loh+GVQKBgQC62CG15Cdf0f26D4KNLMyUfDzB3UgdJPJVuOFrKcpKXasSjDbTc+EX2uxxUSUd2+h8++vNjCvhs2m/sGbHukIHrpR/MT++uE05sWK4JmFGzGurxu9nqS9WwXI9XBx+B8MvhvBSwLt8lIEwnBqMVdxy4P6t2M2OwPJEUlzoU5nfswKBgQC2WsvFyFa0R9MhkMbkz0XAWB4Iu1YEouhLBXfhoSiZkKYLaIo8unVGROMSDNXEz0LaqZFF8/1RIB0IbS3iUirHgEZyLIQvC9fpmABUEe6HIyj4juZYuhZBncQ8vt2Wjxt8FALxx1Fa1ynIR2areazNiW3KH2jeP+cUXMfpe3t2KwKBgEmBtClx/AeXfqGPboYJ5OZZxjFi0/cbTPdqh8x4IWyGU0I1xXAE+7490511FgwcMufQ2GECT1U5F1ZhmN3kqguJpEQx5Oksar1Sywq1lrmavJVU62S1y7ju2/nF3jO67BArnyp/RoNpjTXJhCxHrzXGzIsqaxxJTkaQvJpuIXA/AoGBAJ1S1i0Lu2oL8WYq/r1W6YmZPEgyP3L+jUR0Mkox/NIIDokXJvRftV/rfLAM7LzAR6BY3OGqL6k2+HVVpFl2pDu8Ooq0R1JDeIKqxdXCJrTmK6nNt4NjAGKZ1zzFOm1zh7XTmfq4CENEEGMe3sAf2Gr3HwZbdOER5q4Voc+Y3hpDAoGBAKdPKiRzC7/xyYIsZYr3tTztiNfbSww5vx2NAlNJS1LxiSRZ93OEcvFXODLrfq0UBeCgQl9WcSRULG1/H/d4K5au1Bg4eQDZoWFnVdHyduIdrkqL8OQCODqF1FITBXbkGRjO6tkqP38Sz+eDdkvXIrW3BcniVvBB3bF0nrEd5GHt`,
		PublicKey:  `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgJEriCLy2On2TkaR8qcTIwFVMomuJERp1ETRwaKsFkDOyHEezro+TIvX8rzH3q0lqLtwGfOcXhFqrbCpACwtRA5CwaqT6PZtNIxIoY6zRThMEJQZBTI+iCvmjp7WYGDnQAJeMPQUYIJvAemEa4uYomVVcyiUGF+h8SirZFY8QnOp/HNx4CroXew//lkP0CGUmpi5wAG9TKOcESPVOLIIJmSQddZGVGtitxDIOCRV/ho8ymqFDioahflCk1szEcE8ZL7sZEtZsxZd343fwAIHzRrweuwIlSChK4MuRJmCKm+twHnwIavbkW4U/3jA2wgVM7ewR7PS8YmKLjwiSGCrSwIDAQAB`,
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestOrderRefund(t *testing.T) {
	p := getPayermax(t)
	res, err := p.OrderRefund(&payermax.OrderRefundReq{
		OutRefundNo:     "HB00405T2412311155264017-ff",
		RefundAmount:    5,
		RefundCurrency:  "SAR",
		OutTradeNo:      "HB00405T2412311155264017",
		RefundNotifyUrl: "https://api-test2.thechatie.com/withoutsign/payment/payermax/notify",
	})
	fmt.Println(res, err)
}
