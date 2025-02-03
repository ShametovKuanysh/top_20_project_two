package main

import (
	"github.com/ShametovKuanysh/top_20_project_two/api"
	"github.com/gin-gonic/gin"
)

func main() {
	api.InitDB()

	r := gin.Default()

	r.POST("/register", api.Register)
	r.POST("/login", api.Login)

	protected := r.Group("/tasks/", api.JWTAuthMiddleware())
	protected.GET("/", api.GetTasks)
	protected.POST("/", api.CreateTask)
	protected.GET("/:id", api.GetTask)
	protected.PUT("/:id", api.UpdateTask)
	protected.DELETE("/:id", api.DeleteTask)

	r.Run(":8080")
}
