package handlers

import (
	"agri-api/internal/auth"
	"agri-api/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	notifService *usecase.NotificationService
	jwtService   *auth.JWTService
	blacklist    *auth.Blacklist
}

func NewWebSocketHandler(notifService *usecase.NotificationService, jwtService *auth.JWTService, blacklist *auth.Blacklist) *WebSocketHandler {
	return &WebSocketHandler{
		notifService: notifService,
		jwtService:   jwtService,
		blacklist:    blacklist,
	}
}

func (h *WebSocketHandler) HandleNotifications(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := h.jwtService.Parse(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	jti, _ := claims["jti"].(string)
	if jti != "" {
		if blacklisted, _ := h.blacklist.IsBlacklisted(c.Request.Context(), jti); blacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		}
	}

	userIDStr, _ := claims["user_id"].(string)
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	sub := h.notifService.Subscribe(userID)
	defer h.notifService.Unsubscribe(sub)

	// keep-alive ping
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// drain incoming messages (client pong/close)
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for notification := range sub.Ch {
		data, err := json.Marshal(notification)
		if err != nil {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
