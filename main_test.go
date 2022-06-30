package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12/httptest"

	"github.com/Prep50mobileApp/prep50-api/cmd/prep50"
)

var app *prep50.Prep50

func init() {
	app = prep50.NewApp()
	app.RegisterAppRoutes()
	app.RegisterMiddlewares()
	app.AuthConfig()
	app.RegisterStructValidation()
	pr := exec.Command("go run", "")
	fmt.Println(pr.Run())
}

func TestRegister(t *testing.T) {
	e := httptest.New(t, app.App)
	e.POST("/auth/register").WithJSON(map[string]interface{}{
		"username": "epikoder",
		"password": "password",
		"email":    "efedua.bell@gmail.com",
		"phone":    "09052257844",
	}).Expect().Body().Contains("success")
}

func TestLogin(t *testing.T) {
	e := httptest.New(t, app.App)
	e.POST("/auth/login").WithHeader("Content-type", "application/json").WithJSON(map[string]interface{}{
		"username":    "epikoder",
		"password":    "beLL1923",
		"device_name": "samsung note 8",
		"device_id":   uuid.New().String(),
	}).Expect().Status(http.StatusOK).Body().Contains("success")
}
