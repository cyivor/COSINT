package handlers

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cyivor/cosint/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// localhost:x/
func RootHandler(capir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Redirecting",
			"capir": capir,
		})
	}
}

// localhost:x/home.tmpl
func HomeHandler(capir string, extapir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title":   "Dashboard",
			"extapir": extapir,
			"capir":   capir,
		})
	}
}

// localhost:x/<key>/cosint/identity
func VerifyIdentity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "placeholder"})
}

// localhost:x/<key>/cosint/ext-apis/snusbase GET
func SnusHandler(extapir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "snusbase.tmpl", gin.H{
			"title":   "Snusbase search",
			"extapir": extapir,
		})
	}
}

// localhost:x/<key>/cosint/ext-apis/snusbase POST
func SnusResults(capir string, snusKey string, ratelimitValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if snusKey == "" {
			c.HTML(http.StatusUnauthorized, "enverror.tmpl", gin.H{
				"title": "Environment Variable Error",
				"key":   "SNUSBASE_KEY",
			})
			time.Sleep(5 * time.Second)
			c.Redirect(http.StatusFound, capir)
			return
		}

		searchTerm := c.PostForm("search")
		field := c.PostForm("field")

		logger := c.MustGet("logger").(*zap.Logger)

		if searchTerm == "" || field == "" {
			logger.Error("Invalid form data", zap.String("search", searchTerm), zap.String("field", field))
			c.JSON(http.StatusBadRequest, gin.H{"error": "search and field parameters are required"})
			return
		}

		searchBody := types.RequestBody{
			Terms: []string{searchTerm},
			Types: []string{field},
		}

		response, err := sendRequest("data/search", snusKey, searchBody)
		if err != nil {
			logger.Error("Failed to send Snusbase API request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Snusbase API request failed: %v", err)})
			return
		}

		/*
			logger.Info("Snusbase API request successful",
				zap.String("search", searchTerm),
				zap.String("field", field),
				zap.Any("response", response))
		*/

		RatelimitCount, err := AddLocalRL("_sb") // load ratelimit value
		if err != nil {
			c.HTML(http.StatusUnauthorized, "generalerror.tmpl", gin.H{
				"title": "AddLocalRL(\"_sb\")",
				"err":   err,
			})
			time.Sleep(5 * time.Second)
			c.Redirect(http.StatusFound, capir)
			return
		}
		if RatelimitCount < ratelimitValue {
			c.JSON(http.StatusOK, response)
			return
		}
		c.HTML(http.StatusTooManyRequests, "generalerror.tmpl", gin.H{
			"title": "You are ratelimited",
			"err":   "You have sent too many requests. This is a local ratelimit put in place so you don't abuse your Snusbase API key. If you would like to change it, navigate to your COSINT directory, use a text editor to modify the .env file, and change the value of SBRATELIMIT. The maximum requests you can send to Snusbase in 12 hours is 2,048.",
		})
		time.Sleep(5 * time.Second)
		c.Redirect(http.StatusFound, capir)
	}
}

func RLResponse(api string) string {
	rlMap := map[string]types.LocalRateLimits{
		"_sb": {Snusbase: "sbrl"},
		"_ns": {NoSINT: "nsrl"},
	}

	// get the rate limit struct for the api passed
	rl, exists := rlMap[api]
	if !exists {
		log.Fatalf("no rate limit configuration found for: %s", api)
		// return api // uncomment for error handling instead of fatal
	}

	var rlapi string
	switch api {
	case "_sb":
		rlapi = rl.Snusbase
	case "_ns":
		rlapi = rl.NoSINT
	default:
		log.Fatalf("unknown api: %s", api)
		// return api // uncomment for error handling instead of fatal
	}
	return rlapi
}

func AddLocalRL(api string) (int, error) {
	rlapi := RLResponse(api)
	cachefile := ".rl/" + rlapi

	file, err := os.OpenFile(cachefile, os.O_RDWR, 0600)
	if err != nil {
		log.Fatalf("can't open file. Err: %v", err)
		// return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		log.Fatal("file is empty")
		// return 0, err
	}
	firstLine := scanner.Text()

	num, err := strconv.Atoi(firstLine)
	if err != nil {
		log.Fatalf("%v has been modified manually. Err: %v", file, err)
		// return 0, err
	}

	num++

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatalf("couldn't seek file. Err %v", err)
		// return num, err
	}

	err = file.Truncate(0)
	if err != nil {
		log.Fatalf("couldn't truncate file. Err %v", err)
		// return num, err
	}

	_, err = file.WriteString(strconv.Itoa(num) + "\n// Do not modify this file. It will prevent you from being able to send requests")
	if err != nil {
		log.Fatalf("can't write to %v. Err: %v", file, err)
		// return num, err
	}
	return num, nil
}
