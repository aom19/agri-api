package handlers

import (
	"agri-api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	service *usecase.RBACService
}

func NewRBACHandler(service *usecase.RBACService) *RBACHandler {
	return &RBACHandler{service: service}
}

// ─── Roles ────────────────────────────────────────────────────────────────────

// Rolurile sunt fixe (admin, manager, operator, viewer), definite în migrația
// 000054_fixed_roles_and_permissions, deci API-ul doar le citește.

type assignRoleRequest struct {
	RoleID int64 `json:"role_id" binding:"required"`
}

func (h *RBACHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.service.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (h *RBACHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	role, err := h.service.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

// ─── Permissions pe rol ───────────────────────────────────────────────────────

func (h *RBACHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	perms, err := h.service.GetRolePermissions(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// ─── Permissions list ─────────────────────────────────────────────────────────

func (h *RBACHandler) GetAllPermissions(c *gin.Context) {
	perms, err := h.service.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

func (h *RBACHandler) GetPermission(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	perm, err := h.service.GetPermissionByID(id)
	if err != nil {
		if err.Error() == "permission not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, perm)
}

// ─── My permissions ─────────────────────────────────────────────────────────

// GetMyPermissions returnează permisiunile rolului utilizatorului autentificat.
// Nu necesită nicio permisiune specială — doar un token valid.
func (h *RBACHandler) GetMyPermissions(c *gin.Context) {
	roleIDRaw, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing role in context"})
		return
	}
	roleIDFloat, ok := roleIDRaw.(float64)
	if !ok || roleIDFloat <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid role in token"})
		return
	}
	perms, err := h.service.GetRolePermissions(int64(roleIDFloat))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, perms)
}

// ─── Assign role to user ──────────────────────────────────────────────────────

func (h *RBACHandler) AssignRoleToUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req assignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AssignRoleToUser(userID, req.RoleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role assigned, user sessions revoked"})
}
