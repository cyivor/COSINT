package main

/*
 * author - github.com/cyivor
 * short-term todo
    * create user database
      * encrypted

 * long-term todo
    * add support for nosint
    * add support for maigret
    * add support for telegram bots
      * parsing modules will be needed
    * add dorking support
*/

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cyivor/cosint/handlers"
	"cyivor/cosint/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func RAPIRK() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

func main() {
	logger, err := logger.NewLogger()

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	key, err := RAPIRK()
	if err != nil {
		logger.Fatal("Failed to generate key", zap.Error(err))
	}
	apiRoute := "/" + key

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		logger.Fatal("JWT_SECRET environment variable not set")
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// root route
	r.GET("/", handlers.RootHandler(apiRoute))

	// auth routes
	r.GET("/auth", handlers.AuthHandler)
	r.POST("/login", handlers.LoginHandler(apiRoute, jwtSecret))

	// protected cosint routes
	cosint := r.Group(apiRoute+"/cosint", handlers.AuthMiddleware(apiRoute, jwtSecret, logger))
	{
		cosint.GET("/", handlers.HomeHandler(apiRoute))
		cosint.GET("/identity", handlers.VerifyIdentity)
	}

	// http server
	s := &http.Server{
		Addr:           ":8000",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info("COSINT is online",
		zap.String("url", "http://127.0.0.1:8000"),
		zap.String("api_route", fmt.Sprintf("http://localhost:8000%s/cosint", key)),
	)
	err = s.ListenAndServe()
	if err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
