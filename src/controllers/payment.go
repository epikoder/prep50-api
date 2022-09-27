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
		Action    string
		Type      string `validate:"required_if=id"`
		Provider  string
		Reference string `validate:"required" `
		Id        string
	}
	data := &paymentData{}
	if err := ctx.ReadJSON(data); err != nil {
		ctx.JSON(validation.Errors(err))
		ctx.StatusCode(400)
		return
	}
	fmt.Println(data)

	provider := payment.New(data.Provider)
	res, err := provider.IVerify(data.Reference)
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
			if !ok || tx.Status != "success" {
				ctx.JSON(apiResponse{
					"status": "failed",
				})
				return
			}
		}
	}
	ctx.JSON(apiResponse{
		"status":  "success",
		"message": "Transaction successful",
	})
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
