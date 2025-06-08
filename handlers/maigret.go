package handlers

import (
	"cyivor/cosint/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

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

	report, err := FetchReport(user)
	if err != nil {
		log.Fatalf("couldn't fetch report. Err: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(report), &result)
	if err != nil {
		log.Fatalf("error parsing json. Err %v", err)
	}

	// holds structs
	maigretList := []types.Maigret{}

	// each top level key
	for _, siteData := range result {
		siteMap, ok := siteData.(map[string]interface{})
		if !ok {
			log.Printf("skipping invalid site data: not a map")
			continue
		}

		// fields for struct
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
