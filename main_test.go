package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12/httptest"

	"github.com/Prep50mobileApp/prep50-api/cmd/prep50"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
)

var app *prep50.Prep50

func init() {
	app = prep50.NewApp()
	app.RegisterAppRoutes()
	app.RegisterMiddlewares()
	app.AuthConfig()
	app.RegisterStructValidation()
	ijwt.InitializeSigners()
	go queue.Run() // Use Queue
	generateValues()
}

var (
	username, password, email, phone string
)

func generateValues() {
	username, password = strings.ToLower(crypto.Random(12)), "password"
	email, phone = fmt.Sprintf("%s@gmail.com", strings.ToLower(crypto.Random(12))), fmt.Sprintf("09052257%d%d%d", crypto.RandomNumber(0, 9), crypto.RandomNumber(0, 7), crypto.RandomNumber(4, 9))
}

func TestRegisterFailed(t *testing.T) {
	e := httptest.New(t, app.App)
	e.POST("/auth/register").WithJSON(map[string]interface{}{
		"username": username,
		"password": password,
		"email":    "invalidEmail.com",
		"phone":    "09000000000",
	}).Expect().Status(400).Body().Contains("failed")
}

func TestRegisterSuccess(t *testing.T) {
	e := httptest.New(t, app.App)
	resp := e.POST("/auth/register").WithJSON(map[string]interface{}{
		"username": username,
		"password": password,
		"email":    email,
		"phone":    phone,
	}).Expect().Body()
	resp.Contains("success")
}

func TestLoginFailed(t *testing.T) {
	e := httptest.New(t, app.App)
	e.POST("/auth/login").WithJSON(map[string]interface{}{
		"username": username,
	}).Expect().Status(http.StatusBadRequest).Body().Contains("error")
}

func TestLoginMissingDeviceInfo(t *testing.T) {
	e := httptest.New(t, app.App)
	e.POST("/auth/login").WithJSON(map[string]interface{}{
		"username": username,
		"password": password,
	}).Expect().Status(http.StatusForbidden).Body().Contains("400")
}

var deviceName, deviceID = crypto.Random(12), uuid.New()

func TestLoginSuccess(t *testing.T) {
	generateValues()

	TestRegisterSuccess(t)
	e := httptest.New(t, app.App)
	e.POST("/auth/login").WithJSON(map[string]interface{}{
		"username":    username,
		"password":    password,
		"device_name": deviceName,
		"device_id":   deviceID,
	}).Expect().Status(http.StatusOK).Body().Contains("success")
}

func TestLoginFailedOnNewDevice(t *testing.T) {
	generateValues()
	TestRegisterSuccess(t)
	deviceName, deviceID = crypto.Random(12), uuid.New()
	TestLoginSuccess(t)

	deviceName, deviceID = crypto.Random(12), uuid.New()
	e := httptest.New(t, app.App)
	e.POST("/auth/login").WithJSON(map[string]interface{}{
		"username":    username,
		"password":    password,
		"device_name": deviceName,
		"device_id":   uuid.New(),
	}).Expect().Status(http.StatusForbidden).Body().Contains("failed")
}

func TestLoginFailedOnAnotherUserDevice(t *testing.T) {
	generateValues()
	TestRegisterSuccess(t)
	deviceName, deviceID = crypto.Random(12), uuid.New()
	TestLoginSuccess(t)

	generateValues()
	TestRegisterSuccess(t)
	e := httptest.New(t, app.App)
	resp := e.POST("/auth/login").WithJSON(map[string]interface{}{
		"username":    username,
		"password":    password,
		"device_name": deviceName, // existing user deviceName
		"device_id":   deviceID,   // existing user deviceId
	}).Expect().Status(http.StatusForbidden).Body()
	resp.Contains("failed")
}

func TestSettingPaystackKey(t *testing.T) {
	modes := []string{"live", "sandbox"}
	for _, m := range modes {
		var def = ""
		if m == "live" {
			def = os.Getenv("PAYSTACK_KEY")
		}
		v := settings.GetString("paystack."+m+".secretKey", def)
		if len(v) == 0 {
			t.Error("Expecting sk_*****: Got ''")
		}
		if m == "sandbox" {
			m = "test"
		}
		var vs []string
		if vs = strings.Split(v, "_"); len(vs) != 3 || vs[1] != m {
			t.Error("Invalid paystack key")
		}
	}
}

// Test Resources
func TestGetSubjects(t *testing.T) {
	e := httptest.New(t, app.App)
	var token, _ = ijwt.GenerateToken(&models.User{
		UserName: username,
		Email:    email,
	}, username)
	e.GET("/resources/subjects").WithHeader("Authorization", token.Access).Expect().Status(200).Body().Contains("data")
}

func TestExamWithBoth(t *testing.T) {
	e := httptest.New(t, app.App)
	ex := e.GET("/resources/exams").Expect()
	ex.Body().Contains("BOTH")
}
