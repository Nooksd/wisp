package handler

import (
	"net/http"
	"wisp/src/service"

	"github.com/gin-gonic/gin"
)

type ContactHandler struct {
	svc *service.ContactService
}

func NewContactHandler(s *service.ContactService) *ContactHandler {
	return &ContactHandler{svc: s}
}

func (h *ContactHandler) GetContacts(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	lst, err := h.svc.GetContacts(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lst)
}

func (h *ContactHandler) GetIncoming(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	lst, err := h.svc.GetIncomingRequests(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lst)
}

func (h *ContactHandler) GetSent(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	lst, err := h.svc.GetSentRequests(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lst)
}

func (h *ContactHandler) SendRequest(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	var body struct {
		ToUserID string `json:"toUserId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.svc.SendFriendRequest(c.Request.Context(), uid, body.ToUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"requestId": id.Hex()})
}

func (h *ContactHandler) CancelRequest(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	if err := h.svc.CancelFriendRequest(c.Request.Context(), c.Param("id"), uid); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ContactHandler) AcceptRequest(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	if err := h.svc.AcceptFriendRequest(c.Request.Context(), c.Param("id"), uid); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ContactHandler) RejectRequest(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	if err := h.svc.RejectFriendRequest(c.Request.Context(), c.Param("id"), uid); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ContactHandler) RemoveContact(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	if err := h.svc.RemoveContact(c.Request.Context(), uid, c.Param("id")); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
