package routes

import (
	"wisp/src/handler"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(secure *gin.RouterGroup, public *gin.RouterGroup, h *handler.AuthHandler) {
	public.POST("/auth/register", h.Register)
	public.POST("/auth/login", h.Login)
	secure.POST("/auth/logout", h.Logout)
}
