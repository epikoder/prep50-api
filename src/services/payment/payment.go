package payment

type (
	Pay interface {
		ICharge()
	}

	PaymentConfig struct {
		Paystack map[string]ipaystackConfig
	}
)

func New(method string) Pay {
	switch method {
	default:
		return newIPaystack()
	}
}
