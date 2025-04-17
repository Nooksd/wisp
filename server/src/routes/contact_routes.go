package routes

import (
	"wisp/src/handler"

	"github.com/gin-gonic/gin"
)

func ContactRoutes(secure *gin.RouterGroup, h *handler.ContactHandler) {
	secure.GET("/contacts", h.GetContacts)
	secure.GET("/contacts/requests", h.GetIncoming)
	secure.GET("/contacts/requests/sent", h.GetSent)
	secure.POST("/contacts/requests", h.SendRequest)
	secure.DELETE("/contacts/requests/:id", h.CancelRequest)
	secure.POST("/contacts/requests/:id/accept", h.AcceptRequest)
	secure.POST("/contacts/requests/:id/reject", h.RejectRequest)
	secure.DELETE("/contacts/:id", h.RemoveContact)
}
