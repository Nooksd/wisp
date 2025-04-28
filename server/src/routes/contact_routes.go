package routes

import (
	"wisp/src/handler"

	"github.com/gin-gonic/gin"
)

func ContactRoutes(secure *gin.RouterGroup, h *handler.ContactHandler) {
	contacts := secure.Group("/contacts")
	{
		contacts.GET("", h.GetContacts)
		contacts.GET("/requests", h.GetIncoming)
		contacts.GET("/requests/sent", h.GetSent)
		contacts.POST("/requests", h.SendRequest)
		contacts.DELETE("/requests/:id", h.CancelRequest)
		contacts.POST("/requests/:id/accept", h.AcceptRequest)
		contacts.POST("/requests/:id/reject", h.RejectRequest)
		contacts.DELETE("/:id", h.RemoveContact)
	}
}
