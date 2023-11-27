package payment

import (
	"net/http"
	"os"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/epikoder/paystack-go"
)

type (
	ipaystack struct {
		paystack.Client
	}
)

func newIPaystack() *ipaystack {
	mode := settings.GetString("paystack.mode", "live")

	return &ipaystack{
		Client: *paystack.NewClient(settings.GetString("paystack."+mode+".secretKey", os.Getenv("PAYSTACK_KEY")), &http.Client{}),
	}
}

func (p *ipaystack) Name() string {
	return "paystack"
}

func (p *ipaystack) ICharge(req ChargeRequest) (interface{}, error) {
	return p.Client.Charge.Create(&paystack.ChargeRequest{
		Email:             req.Email,
		Amount:            req.Amount * 100,
		Birthday:          req.Birthday,
		Metadata:          (*paystack.Metadata)(req.Metadata),
		AuthorizationCode: req.AuthorizationCode,
		Pin:               req.Pin,
	})
}

func (p *ipaystack) IVerify(reference string) (interface{}, error) {
	return p.Client.Transaction.Verify(reference)
}

func (p *ipaystack) Initialize(req TransactionRequest) (interface{}, error) {
	tx := &paystack.TransactionRequest{
		CallbackURL: req.CallbackURL,
		Currency:    req.Currency,
		Amount:      req.Amount * 100,
		Email:       req.Email,
		Metadata:    paystack.Metadata(req.Metadata),
		Reference:   req.Reference,
	}
	return p.Client.Transaction.Initialize(tx)
}
