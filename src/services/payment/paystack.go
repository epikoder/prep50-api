package payment

import (
	"github.com/rpip/paystack-go"
)

type (
	ipaystack struct {
		paystack.Client
	}

	ipaystackConfig struct {
		PkKey string `yaml:"pk_key"`
		SkKey string `yaml:"sk_key"`
	}
)

func newIPaystack() *ipaystack {
	return &ipaystack{}
}

func (p *ipaystack) ICharge() {
	p.Client.Charge.Create(&paystack.ChargeRequest{})
}
