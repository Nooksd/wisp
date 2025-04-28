package routes

import (
	"wisp/src/handler"

	"github.com/gin-gonic/gin"
)

func WSRoutes(secure *gin.RouterGroup, wsHandler *handler.WSHandler) {
	ws := secure.Group("/ws")
	{
		ws.GET("", wsHandler.HandleConnection)
	}
}
