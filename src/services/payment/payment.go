package payment

type (
	PaymentProvider interface {
		ICharge()
		IVerify(string) (interface{}, error)
		Name() string
	}
)

func New(provider string) PaymentProvider {
	switch provider {
	default:
		return newIPaystack()
	}
}
