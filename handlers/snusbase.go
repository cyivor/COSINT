package handlers

import (
	"cyivor/cosint/types"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
		c.Writer.Header()
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
