package main

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12/httptest"

	"github.com/Prep50mobileApp/prep50-api/cmd/prep50"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/crypto"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
)

var app *prep50.Prep50

func init() {
	app = prep50.NewApp()
	app.RegisterAppRoutes()
	app.RegisterMiddlewares()
	app.AuthConfig()
	app.RegisterStructValidation()
	go queue.Run() // Use Queue
}

var username, password = crypto.Random(12), "password"

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
		"email":    "efedua.bell@gmail.com",
		"phone":    "09000000000",
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
	e.GET("/resources/subjects").Expect().Status(200).Body().Contains("data")
}
