package handlers

import (
	"agri-api/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *usecase.AuthService
}

func NewAuthHandler(authService *usecase.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login autentifică un utilizator
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{email=string,password=string} true "Credențiale"
// @Success      200 {object} object{access_token=string,refresh_token=string}
// @Failure      400 {object} object{errors=object}
// @Failure      401 {object} object{error=string}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	access, refresh, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})

}

// Refresh rotează refresh token-ul și emite un access token nou
// @Summary      Refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization header string false "Bearer <access_token_vechi>"
// @Param        body body object{refresh_token=string} true "Refresh token"
// @Success      200 {object} object{access_token=string,refresh_token=string}
// @Failure      400 {object} object{errors=object}
// @Failure      401 {object} object{error=string}
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	// Extrage access token-ul vechi din header pentru blacklisting
	oldAccessToken := ""
	if authHeader := c.GetHeader("Authorization"); len(authHeader) > 7 {
		oldAccessToken = authHeader[7:]
	}

	newAccess, newRefresh, err := h.authService.Refresh(req.RefreshToken, oldAccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": newAccess, "refresh_token": newRefresh})
}

// Logout revocă sesiunea curentă
// @Summary      Logout
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body object{refresh_token=string} true "Refresh token"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{errors=object}
// @Failure      500 {object} object{error=string}
// @Router       /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	// Extrage access token-ul din header pentru blacklisting
	accessToken := ""
	if authHeader := c.GetHeader("Authorization"); len(authHeader) > 7 {
		accessToken = authHeader[7:]
	}

	if err := h.authService.Logout(req.RefreshToken, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not logout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// Register înregistrează un utilizator nou
// @Summary      Register
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{email=string,password=string,role=string} true "Date utilizator"
// @Success      201 {object} object{access_token=string,refresh_token=string}
// @Failure      400 {object} object{error=string}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	access, refresh, err := h.authService.Register(req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"access_token": access, "refresh_token": refresh})
}

// ForgotPassword inițiază resetarea parolei
// @Summary      Forgot password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body object{email=string} true "Email utilizator"
// @Success      200 {object} object{message=string,reset_token=string}
// @Failure      400 {object} object{errors=object}
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	token, err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process request"})
		return
	}
	// Răspuns generic indiferent dacă email-ul există (securitate)
	resp := gin.H{"message": "if the email exists, a reset link was sent"}
	if token != "" {
		// TODO: în producție trimite token-ul pe email, nu în răspuns
		resp["reset_token"] = token
	}
	c.JSON(http.StatusOK, resp)
}

// ResetPassword resetează parola cu token-ul primit
// @Summary      Reset password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token path string true "Token de resetare"
// @Param        body body object{password=string,confirm_password=string} true "Parola nouă"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Router       /auth/reset-password/{token} [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	var req struct {
		Password        string `json:"password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors(err)})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"errors": gin.H{"confirm_password": "passwords do not match"}})
		return
	}

	if err := h.authService.ResetPassword(token, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}
