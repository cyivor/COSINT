package handlers

import (
	"cyivor/cosint/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// localhost:x/<key>/cosint/int-apis/maigret GET
func MaigretHandler(intapir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "maigret.tmpl", gin.H{
			"title":   "Maigret search",
			"intapir": intapir,
		})
	}
}

// localhost:x/<key>/cosint/int-apis/maigret POST
func MaigretResults(capir string, ratelimitValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		searchTerm := c.PostForm("search")
		logger := c.MustGet("logger").(*zap.Logger)

		if searchTerm == "" {
			logger.Error("Invalid form data", zap.String("search", searchTerm))
			c.JSON(http.StatusBadRequest, gin.H{"error": "search parameter is required"})
			return
		}

		// maigret results
		response := ParseReport(searchTerm)

		// check rate limit
		RatelimitCount, err := AddLocalRL("_mg")
		if err != nil {
			logger.Error("Rate limit check failed", zap.Error(err))
			c.HTML(http.StatusUnauthorized, "generalerror.tmpl", gin.H{
				"title": "Rate Limit Error",
				"err":   err.Error(),
			})
			time.Sleep(5 * time.Second)
			c.Redirect(http.StatusFound, capir)
			return
		}

		if RatelimitCount >= ratelimitValue {
			logger.Warn("Rate limit exceeded", zap.Int("count", RatelimitCount))
			c.HTML(http.StatusTooManyRequests, "generalerror.tmpl", gin.H{
				"title": "You are ratelimited",
				"err":   "You have sent too many requests. This is a local ratelimit put in place so you aren't barred from sending requests to the sites Maigret matches usernames against. If you would like to change it, navigate to your COSINT directory, use a text editor to modify the .env file, and change the value of MGRATELIMIT.",
			})
			time.Sleep(5 * time.Second)
			c.Redirect(http.StatusFound, capir)
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// maigret --no-color --no-progressbar -J=simple -a --id-type=username userhere
func MaigretCommand(user string) {
	exec.Command("maigret", "--no-color", "--no-progressbar", "-J", "simple", "-a", "--id-type", "username", "-n", "25", user).Output()
}

// pythonlibs/maigret/reports/report_userhere_simple.json
func FetchReport(user string) (string, error) {
	fsep := string(filepath.Separator)
	report := "pythonlibs" + fsep + "maigret" + fsep + "reports" + fsep + "report_" + user + "_simple.json"
	reportInfo, err := os.Lstat(report)
	if err != nil {
		fmt.Println(err)
		return report, err
	}
	mode := reportInfo.Mode()
	if !mode.IsRegular() {
		return report, fmt.Errorf("report for %s not found", user)
	}
	return report, nil
}

func ParseReport(user string) []types.Maigret {
	MaigretCommand(user)

	reportPath, err := FetchReport(user)
	if err != nil {
		log.Fatalf("couldn't fetch report. Err: %v", err)
	}

	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		log.Fatalf("couldn't read report file %s. Err: %v", reportPath, err)
	}

	// parse json into a map
	var result map[string]interface{}
	err = json.Unmarshal(reportData, &result)
	if err != nil {
		log.Fatalf("error parsing JSON. Err: %v", err)
	}

	maigretList := []types.Maigret{}

	// top level keys
	for _, siteData := range result {
		siteMap, ok := siteData.(map[string]interface{})
		if !ok {
			log.Printf("skipping invalid site data: not a map")
			continue
		}

		maigret := types.Maigret{}

		// url_user
		if urlUser, ok := siteMap["url_user"].(string); ok {
			maigret.UrlUser = urlUser
		}

		// username
		if username, ok := siteMap["username"].(string); ok {
			maigret.User = username
		}

		// site object
		if site, ok := siteMap["site"].(map[string]interface{}); ok {
			// tags
			if tags, ok := site["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						maigret.Tags = append(maigret.Tags, tagStr)
					}
				}
			}
		}

		// status object
		if status, ok := siteMap["status"].(map[string]interface{}); ok {
			// site_name
			if siteName, ok := status["site_name"].(string); ok {
				maigret.SiteName = siteName
			}
		}

		maigretList = append(maigretList, maigret)
	}

	return maigretList
}
