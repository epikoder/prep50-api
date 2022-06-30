package main

import (
	"github.com/Prep50mobileApp/prep50-api/cmd/prep50"
	"github.com/Prep50mobileApp/prep50-api/src/services/database/queue"
)

func main() {
	prep50 := prep50.NewApp()
	prep50.RegisterAppRoutes()
	prep50.RegisterMiddlewares()
	prep50.RegisterStructValidation()
	prep50.AuthConfig()
	go queue.Run()
	prep50.StartServer()
}
