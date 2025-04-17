package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckAdminOrUidPermission(c *gin.Context, targetUid string) bool {
	isAdminVal, exists := c.Get("isAdmin")
	isAdmin := exists && isAdminVal.(bool)

	userIdVal, exists := c.Get("userId")
	userId := ""
	if exists {
		userId = userIdVal.(string)
	}

	if !isAdmin && userId != targetUid {
		c.JSON(http.StatusForbidden, gin.H{"error": "Você não tem permissão para acessar este recurso"})
		return false
	}
	return true
}
