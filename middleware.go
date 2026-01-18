package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check credentials against database
		var passwordHash string
		err := db.QueryRow("SELECT password_hash FROM admin_users WHERE username = ?", username).Scan(&passwordHash)
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Compare password with hash
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("username", username)
		c.Next()
	}
}
