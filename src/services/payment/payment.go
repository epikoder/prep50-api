package payment

type (
	Metadata      map[string]interface{}
	ChargeRequest struct {
		Email             string    `json:"email,omitempty"`
		Amount            float32   `json:"amount,omitempty"`
		Birthday          string    `json:"birthday,omitempty"`
		AuthorizationCode string    `json:"authorization_code,omitempty"`
		Pin               string    `json:"pin,omitempty"`
		Metadata          *Metadata `json:"metadata,omitempty"`
	}

	TransactionRequest struct {
		CallbackURL string   `json:"callback_url,omitempty"`
		Reference   string   `json:"reference,omitempty"`
		Currency    string   `json:"currency,omitempty"`
		Amount      float32  `json:"amount,omitempty"`
		Email       string   `json:"email,omitempty"`
		Metadata    Metadata `json:"metadata,omitempty"`
	}

	PaymentProvider interface {
		ICharge(ChargeRequest) (interface{}, error)
		IVerify(string) (interface{}, error)
		Initialize(TransactionRequest) (interface{}, error)
		Name() string
	}
)

func New(provider *string) PaymentProvider {
	switch provider {
	default:
		return newIPaystack()
	}
}
