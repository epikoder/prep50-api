package main

import (
	"github.com/Prep50mobileApp/prep50-api/cmd/prep50"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/ijwt"
	"github.com/Prep50mobileApp/prep50-api/src/pkg/settings"
	"github.com/Prep50mobileApp/prep50-api/src/services/queue"
)

func main() {
	prep50 := prep50.NewApp()
	prep50.RegisterAppRoutes()
	prep50.RegisterMiddlewares()
	prep50.RegisterStructValidation()
	prep50.AuthConfig()
	ijwt.InitializeSigners()
	settings.SeedSettings()
	go queue.Run()
	prep50.StartServer()
}
