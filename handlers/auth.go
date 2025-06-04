package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func AuthMiddleware(apiRoute string, jwtSecret []byte, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// check for auth cookie
		tokenString, err := c.Cookie("__cosint")
		if err != nil {
			logger.Warn("No __cosint cookie", zap.String("path", c.Request.URL.Path))
			c.Redirect(http.StatusFound, "/auth")
			c.Abort()
			return
		}

		// validate jwt
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

func LoginHandler(apiRoute string, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.PostForm("userid")
		password := c.PostForm("password")

		// test userid & pass for evident reasons
		if userID == "test" && password == "test" {
			// create JWT
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": userID,                           // aubject (user id)
				"iat": time.Now().Unix(),                // issued at
				"exp": time.Now().Add(time.Hour).Unix(), // expires in 1 hour
			})

			tokenString, err := token.SignedString(jwtSecret)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "auth.tmpl", gin.H{
					"title": "Login",
					"error": "Failed to generate token",
				})
				return
			}

			// __cosint cookie with JWT
			c.SetCookie(
				"__cosint",
				tokenString,
				3600, // 1hr
				"/",
				"",
				false, // secure: false for localhost
				true,  // httponly
			)
			c.Redirect(http.StatusFound, apiRoute+"/cosint")
			return
		}

		// Failed login
		c.HTML(http.StatusUnauthorized, "auth.tmpl", gin.H{
			"title": "Login",
			"error": "Invalid credentials",
		})
	}
}
