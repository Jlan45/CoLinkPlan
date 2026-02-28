package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"CoLinkPlan/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("super-secret-colink-dev-key") // Ideally loaded from env later

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func generateToken(prefix string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set("user_id", claims["sub"])
		c.Set("email", claims["email"])
		c.Next()
	}
}

func RegisterHandler(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		apiToken := generateToken("sk-colink")
		clientToken := generateToken("client")

		if err := database.CreateUser(c.Request.Context(), req.Email, string(hashed), apiToken, clientToken); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
	}
}

func LoginHandler(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		u, err := database.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		claims := jwt.MapClaims{
			"sub":   u.ID,
			"email": u.Email,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := t.SignedString(jwtSecret)

		c.JSON(http.StatusOK, gin.H{"token": tokenString, "user": u})
	}
}

func MeHandler(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.GetString("email")
		u, err := database.GetUserByEmail(c.Request.Context(), email)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": u})
	}
}

func NodesHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		hub.mu.RLock()
		defer hub.mu.RUnlock()

		type NodeInfo struct {
			ID              string   `json:"id"`
			MaxParallel     int      `json:"max_parallel"`
			ActiveTasks     int      `json:"active_tasks"`
			SupportedModels []string `json:"supported_models"`
			Penalized       bool     `json:"penalized"`
		}

		nodes := make([]NodeInfo, 0, len(hub.clients))
		for client := range hub.clients {
			if client.MaxParallel == 0 {
				continue // not fully registered
			}

			models := make([]string, 0, len(client.SupportedModels))
			for m := range client.SupportedModels {
				models = append(models, m)
			}

			nodes = append(nodes, NodeInfo{
				ID:              client.ID, // the client token technically
				MaxParallel:     client.MaxParallel,
				ActiveTasks:     client.ActiveTasks,
				SupportedModels: models,
				Penalized:       time.Now().Before(client.PenaltyUntil),
			})
		}

		c.JSON(http.StatusOK, gin.H{"nodes": nodes})
	}
}
