package handlers

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"

	"cyivor/cosint/types"

	"github.com/gin-gonic/gin"
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
