package payment

type (
	PaymentProvider interface {
		ICharge()
		IVerify(string) (interface{}, error)
	}
)

func New(provider string) PaymentProvider {
	switch provider {
	default:
		return newIPaystack()
	}
}
