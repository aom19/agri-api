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

type createRoleRequest struct {
	Code        string `json:"code"        binding:"required"`
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

type updateRoleRequest struct {
	Code        string `json:"code"        binding:"required"`
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

type setPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids" binding:"required"`
}

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

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req createRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.service.CreateRole(req.Code, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.service.UpdateRole(id, req.Code, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteRole(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
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

func (h *RBACHandler) SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req setPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.SetRolePermissions(id, req.PermissionIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permissions updated"})
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
