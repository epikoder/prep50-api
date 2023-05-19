package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12/httptest"

	"github.com/Prep50mobileApp/prep50-api/cmd/prep50"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
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
}

var (
	username, password = strings.ToLower(crypto.Random(12)), "password"
	email, phone       = fmt.Sprintf("%s@gmail.com", strings.ToLower(crypto.Random(12))), fmt.Sprintf("09052257%d%d%d", crypto.RandomNumber(0, 9), crypto.RandomNumber(0, 7), crypto.RandomNumber(4, 9))
)

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
	e.POST("/auth/register").WithJSON(map[string]interface{}{
		"username": username,
		"password": password,
		"email":    email,
		"phone":    phone,
	}).Expect().Body().Contains("success")
}

var deviceName, deviceID = crypto.Random(12), uuid.New()

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

func TestLoginSuccess(t *testing.T) {
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
	e := httptest.New(t, app.App)
	e.POST("/auth/login").WithJSON(map[string]interface{}{
		"username":    username,
		"password":    password,
		"device_name": deviceName,
		"device_id":   uuid.New(),
	}).Expect().Status(http.StatusForbidden).Body().Contains("failed")
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
