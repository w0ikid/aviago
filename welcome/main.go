package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type WhoamiResponse struct {
	Identity struct {
		ID     string `json:"id"`
		Traits struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"traits"`
	} `json:"identity"`
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Достаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized (no bearer token)"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Отправляем запрос в Kratos /sessions/whoami
		req, _ := http.NewRequest("GET", "http://127.0.0.1:4433/sessions/whoami", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized (invalid session)"})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var whoami WhoamiResponse
		if err := json.Unmarshal(body, &whoami); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized (bad whoami response)"})
			c.Abort()
			return
		}

		// ✅ Авторизованный пользователь
		fmt.Printf("Authorized user: %s (%s)\n", whoami.Identity.Traits.Name, whoami.Identity.Traits.Email)

		c.Set("user_id", whoami.Identity.ID)
		c.Set("user_email", whoami.Identity.Traits.Email)
		c.Set("user_name", whoami.Identity.Traits.Name)

		c.Next()
	}
}

func main() {
	r := gin.Default()

	// public
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is a public endpoint 🌍",
		})
	})

	// private
	r.GET("/welcome", authMiddleware(), func(c *gin.Context) {
		userName := c.GetString("user_name")
		userEmail := c.GetString("user_email")

		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome 🚀",
			"name":    userName,
			"email":   userEmail,
		})
	})

	fmt.Println("Server is running on http://127.0.0.1:4455")
	if err := r.Run(":4455"); err != nil {
		panic(err)
	}
}
