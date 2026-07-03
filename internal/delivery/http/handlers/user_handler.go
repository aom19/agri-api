package handlers

import (
	"agri-api/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *usecase.UserService
}

func NewUserHandler(service *usecase.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type createUserRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"omitempty,min=8"`
	RoleID         int64  `json:"role_id" binding:"required"`
	EmailConfirmed bool   `json:"email_confirmed"`
}

type updateUserRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"omitempty,min=8"`
	RoleID         int64  `json:"role_id" binding:"required"`
	EmailConfirmed bool   `json:"email_confirmed"`
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := h.service.GetUserByID(id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	user, err := h.service.CreateUser(req.Email, req.Password, req.RoleID, req.EmailConfirmed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	user, err := h.service.UpdateUser(id, req.Email, req.RoleID, req.EmailConfirmed, req.Password)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteUser(id); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
