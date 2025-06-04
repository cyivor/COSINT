package main

/*
 * author - github.com/cyivor
 * short-term todo
    * create register section

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

	"cyivor/cosint/db"
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
		log.Fatalf("Failed to initialise logger: %v", err)
	}
	defer logger.Sync()

	dbKey := os.Getenv("DB_KEY")
	if dbKey == "" {
		logger.Fatal("DB_KEY environment variable not set")
	}

	snusKey := os.Getenv("SNUSBASE_KEY")

	database, err := db.InitDB("./cosint.db", dbKey, logger)
	if err != nil {
		logger.Fatal("Failed to initialise database", zap.Error(err))
	}
	defer database.Close()

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
	// r.StaticFile("/favicon.ico", "./static/favicon.ico")
	// only keeping this because ill probably convert a png to ico at some point but i just don't know what logo yet lol

	// protected cosint routes
	// these variables make it easier but now i have to go ahead and edit the template files
	capir := apiRoute + "/cosint"
	extapir := capir + "/ext-apis"

	// root route
	r.GET("/", handlers.RootHandler(apiRoute))

	// set logger context for /login
	r.Use(func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	})

	// auth routes
	r.GET("/auth", handlers.AuthHandler)
	r.POST("/login", handlers.LoginHandler(apiRoute, jwtSecret, database))

	cosint := r.Group(capir, handlers.AuthMiddleware(apiRoute, jwtSecret, logger))
	externalAPIs := r.Group(extapir, handlers.AuthMiddleware(apiRoute, jwtSecret, logger))
	{
		// GET
		cosint.GET("/", handlers.HomeHandler(capir))
		cosint.GET("/identity", handlers.VerifyIdentity)
		cosint.GET("/create-new-user", handlers.NewUserHandler(capir))

		// POST
		cosint.POST("/new", handlers.RegisterHandler(jwtSecret, dbKey))
	}
	{
		// GET
		externalAPIs.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, capir)
		})
		externalAPIs.GET("/snusbase", handlers.SnusHandler(extapir))

		// POST
		externalAPIs.POST("/snusbase", handlers.SnusResults(capir, snusKey))
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
		zap.String("api_route", fmt.Sprintf("http://localhost:8000/%s/cosint", key)),
	)
	err = s.ListenAndServe()
	if err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
