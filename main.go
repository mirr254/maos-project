package main

import (
	"maos-cloud-project-api/config"
	"maos-cloud-project-api/router"
	utils "maos-cloud-project-api/utils"

	"github.com/sirupsen/logrus"
)

func main() {

	cfg := config.LoadConfig()

	_, err := config.InitDB(cfg)
	if err != nil {
		logrus.Error("Database Initialization failed", err)
		return
	}
	
	utils.EnsurePlugins()
	// utils.CreateAwsSession()

	r := utils.SetUpRouter()
	router.AuthRoutes(r, cfg)
	router.HealthCheck(r)
	router.ProjectRoutes(r)
	r.Run(":8080")

}
