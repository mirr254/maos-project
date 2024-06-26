package router

import (
	"maos-cloud-project-api/controllers"

	"github.com/gin-gonic/gin"
)

func ProjectRoutes( r *gin.Engine) {
	v1 := r.Group("/api/v1")

	v1.POST("/project", controllers.CreateProject)
	v1.POST("/stack", controllers.CreateStack)
	v1.GET("/stack", controllers.GetStack)
	v1.POST("/stack/destroy", controllers.DeleteStack)

}