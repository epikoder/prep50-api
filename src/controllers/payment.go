package controllers

import (
	"fmt"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/validation"
	"github.com/Prep50mobileApp/prep50-api/src/services/payment"
	"github.com/epikoder/paystack-go"
	"github.com/kataras/iris/v12"
)

func VerifyPayment(ctx iris.Context) {
	type paymentData struct {
		Action   string
		Type     string `validate:"required"`
		Provider string
		Id       string
	}
	data := &paymentData{}
	if err := ctx.ReadJSON(data); err != nil {
		ctx.JSON(validation.Errors(err))
		ctx.StatusCode(400)
		return
	}

	provider := payment.New(data.Provider)
	res, err := provider.IVerify(data.Id)
	if !logger.HandleError(err) {
		ctx.JSON(apiResponse{
			"status": "failed",
		})
		return
	}

	switch data.Provider {
	default:
		{
			tx, ok := res.(*paystack.Transaction)
			if ok {
				ctx.JSON(apiResponse{
					"status": "failed",
				})
				return
			}
			fmt.Println(tx)
		}
	}
}

func HandlePayment(ctx iris.Context) {
	type paymentData struct {
		Action   string
		Type     string `validate:"required"`
		Provider string
		Id       string
	}
	data := &paymentData{}
	if err := ctx.ReadJSON(data); err != nil {
		ctx.JSON(validation.Errors(err))
		ctx.StatusCode(400)
		return
	}

	provider := payment.New(data.Provider)
	switch data.Action {
	case "verify":
		provider.IVerify(data.Id)
	default:
		provider.ICharge()
	}
}

func PaymentHook(ctx iris.Context) {

}
