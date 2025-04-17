package handler

import (
	"net/http"
	"strconv"
	"wisp/src/helpers"
	"wisp/src/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(u *service.UserService) *UserHandler {
	return &UserHandler{userSvc: u}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIdVal, exists := c.Get("userId")
	uid := ""
	if exists {
		uid = userIdVal.(string)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Erro interno do servidor"})
		return
	}

	user, err := h.userSvc.GetByUserID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	uid := c.Param("userId")
	if !helpers.CheckAdminOrUidPermission(c, uid) {
		return
	}

	u, err := h.userSvc.GetUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuário não encontrado"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	if !helpers.CheckAdminOrUidPermission(c, "") {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	q := c.Query("q")
	order := c.DefaultQuery("order", "createdAt")
	sortDesc := c.DefaultQuery("sort", "desc") == "desc"

	users, total, err := h.userSvc.ListUsers(c.Request.Context(), page, limit, q, order, sortDesc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"users": users,
	})
}

func (h *UserHandler) CheckAvailability(c *gin.Context) {
	email := c.Query("email")
	userID := c.Query("userId")

	ok, err := h.userSvc.CheckAvailability(c.Request.Context(), email, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"available": ok})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	uid := c.Param("userId")
	if !helpers.CheckAdminOrUidPermission(c, uid) {
		return
	}
	var body struct {
		Name    string `json:"name"`
		Email   string `json:"email"  binding:"omitempty,email"`
		IsAdmin *bool  `json:"isAdmin"`
	}

	isAdminVal, exists := c.Get("isAdmin")
	isAdmin := exists && isAdminVal.(bool)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.IsAdmin != nil && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Erro ao atualizar o usuário"})
		return
	}

	upd := make(map[string]any)
	if body.Name != "" {
		upd["name"] = body.Name
	}
	if body.Email != "" {
		upd["email"] = body.Email
	}
	if body.IsAdmin != nil {
		upd["isAdmin"] = *body.IsAdmin
	}
	if len(upd) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nenhum campo para atualizar"})
		return
	}
	if err := h.userSvc.UpdateUser(c.Request.Context(), uid, upd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	uid := c.Param("userId")
	if !helpers.CheckAdminOrUidPermission(c, uid) {
		return
	}
	if err := h.userSvc.DeleteUser(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
