package handlers

import (
	"cyivor/cosint/db"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func AuthMiddleware(apiRoute string, jwtSecret []byte, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("__cosint")
		if err != nil {
			logger.Warn("No __cosint cookie", zap.String("path", c.Request.URL.Path))
			c.Redirect(http.StatusFound, "/auth")
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			logger.Warn("Invalid JWT", zap.String("path", c.Request.URL.Path), zap.Error(err))
			c.Redirect(http.StatusFound, "/auth")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Warn("Invalid JWT claims", zap.String("path", c.Request.URL.Path))
			c.Redirect(http.StatusFound, "/auth")
			c.Abort()
			return
		}

		logger.Info("Authorized access",
			zap.String("path", c.Request.URL.Path),
			zap.String("user_id", claims["sub"].(string)),
		)
		c.Next()
	}
}

func AuthHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "auth.tmpl", gin.H{
		"title": "Login",
	})
}

func NewUserHandler(capir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.tmpl", gin.H{
			"title": "Create new user",
			"capir": capir,
		})
	}
}

func RegisterHandler(jwtSecret []byte, dbKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.PostForm("userid")
		password := c.PostForm("password")

		logger := c.MustGet("logger").(*zap.Logger)

		// validate user against database
		valid, err := db.NewUser("./cosint.db", dbKey, logger, userID, password)
		if err != nil {
			logger.Error("Failed to validate user", zap.String("userid", userID), zap.Error(err))
			c.HTML(http.StatusInternalServerError, "auth.tmpl", gin.H{
				"title": "Login",
				"error": "Internal server error",
			})
			return
		}
		if !valid {
			logger.Warn("Invalid credentials", zap.String("userid", userID))
			c.HTML(http.StatusUnauthorized, "auth.tmpl", gin.H{
				"title": "Login",
				"error": "Invalid credentials",
			})
			return
		}
	}
}

func LoginHandler(apiRoute string, jwtSecret []byte, dbl *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.PostForm("userid")
		password := c.PostForm("password")

		logger := c.MustGet("logger").(*zap.Logger)

		// validate user against database
		valid, err := db.ValidateUser(dbl, userID, password, logger)
		if err != nil {
			logger.Error("Failed to validate user", zap.String("userid", userID), zap.Error(err))
			c.HTML(http.StatusInternalServerError, "auth.tmpl", gin.H{
				"title": "Login",
				"error": "Internal server error",
			})
			return
		}
		if !valid {
			logger.Warn("Invalid credentials", zap.String("userid", userID))
			c.HTML(http.StatusUnauthorized, "auth.tmpl", gin.H{
				"title": "Login",
				"error": "Invalid credentials",
			})
			return
		}

		// create jwt
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userID,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			logger.Error("Failed to generate JWT", zap.String("userid", userID), zap.Error(err))
			c.HTML(http.StatusInternalServerError, "auth.tmpl", gin.H{
				"title": "Login",
				"error": "Failed to generate token",
			})
			return
		}

		// set the __cosint cookie with JWT
		c.SetCookie(
			"__cosint",
			tokenString,
			3600,
			"/",
			"",
			false,
			true,
		)
		logger.Info("Successful login", zap.String("userid", userID))
		c.Redirect(http.StatusFound, apiRoute+"/cosint")
	}
}
