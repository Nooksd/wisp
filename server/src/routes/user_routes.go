package routes

import (
	"wisp/src/handler"

	"github.com/gin-gonic/gin"
)

func UserRoutes(secure *gin.RouterGroup, public *gin.RouterGroup, h *handler.UserHandler) {
	public.GET("/check", h.CheckAvailability)
	secure.GET("/me", h.GetProfile)
	secure.GET("/users", h.ListUsers)
	secure.GET("/users/:userId", h.GetUser)
	secure.PUT("/users/:userId", h.UpdateUser)
	secure.DELETE("/users/:userId", h.DeleteUser)
}
