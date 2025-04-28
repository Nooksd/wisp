package routes

import (
	"wisp/src/handler"

	"github.com/gin-gonic/gin"
)

func UserRoutes(secure *gin.RouterGroup, public *gin.RouterGroup, h *handler.UserHandler) {
	public.GET("/check", h.CheckAvailability)
	secure.GET("/me", h.GetProfile)

	users := secure.Group("/users")
	{
		users.GET("", h.ListUsers)
		users.GET("/:userId", h.GetUser)
		users.PUT("/:userId", h.UpdateUser)
		users.DELETE("/:userId", h.DeleteUser)
	}
}
