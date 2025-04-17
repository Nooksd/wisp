package handler

import (
	"net/http"
	"time"
	"wisp/src/model"
	"wisp/src/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
	userSvc *service.UserService
}

func NewAuthHandler(a *service.AuthService, u *service.UserService) *AuthHandler {
	return &AuthHandler{authSvc: a, userSvc: u}
}

func (h *AuthHandler) Register(c *gin.Context) {

	var body model.RegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userSvc.Register(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, u)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
		DeviceID string `json:"deviceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authSvc.Login(c.Request.Context(), body.Email, body.Password, body.DeviceID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    "Barear " + token,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sid := c.GetString("sid")
	if err := h.authSvc.Logout(c.Request.Context(), sid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(1 * time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	c.Status(http.StatusNoContent)
}
