package payment

import (
	"net/http"

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
		Client: *paystack.NewClient(settings.GetString("paystack."+mode+".secretKey", ""), &http.Client{}),
	}
}

func (p *ipaystack) ICharge() {
	p.Client.Charge.Create(&paystack.ChargeRequest{})
}

func (p *ipaystack) IVerify(reference string) (interface{}, error) {
	return p.Client.Transaction.Verify(reference)
}
