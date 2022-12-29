package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/archiver/v3"
)

var backupDir string

const (
	backupDirDefault = "/opt/pmm_grafana_backup_tool/_OUTPUT_"
	limit            = 5000
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// getGrafanaData calls the grafana API endpoint for the provided GRAFANA_URL
func getGrafanaData(grafanaUrlPrefix, grafanaToken, endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", grafanaUrlPrefix+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", grafanaToken))
	req.Header.Set("Content-Type", "application/json")
	params := req.URL.Query()
	params.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = params.Encode()

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Println("response.StatusCode: ", response.StatusCode)
		return nil, fmt.Errorf("[Error] %s", response.Status)
	}

	read, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return read, nil
}

// getGrafanaData calls the grafana API endpoint for the provided GRAFANA_URL
func postGrafanaData(grafanaUrlPrefix, grafanaToken, endpoint string, body any) error {
	bytesBody, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	//fmt.Println("body: ", string(bytesBody))
	req, err := http.NewRequest("POST", grafanaUrlPrefix+endpoint, bytes.NewBuffer(bytesBody))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", grafanaToken))
	req.Header.Set("Content-Type", "application/json")
	params := req.URL.Query()
	params.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = params.Encode()

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		all, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		fmt.Println("error: ", string(all))
		return fmt.Errorf("[Error] %s", response.Status)
	}

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return nil
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// syncDashboards replicates the grafana folder structure, downloads all
// dashbords using the grafana api, and places them in each folder
func syncDashboards(grafanaUrl, grafanaToken string, dashboards dashboardSearch) {
	dashDir := backupDir
	if _, err := os.Stat(dashDir); !os.IsNotExist(err) {
		os.RemoveAll(dashDir)
	}
	//err := os.Mkdir(dashDir, 0755)
	//check(err)

	var failed, total int
	log.Println("Syncing Dashboards...")
	for _, ds := range dashboards {
		if ds.Type == "dash-folder" {
			// check if a folder for exists, if not, create one
			if _, err := os.Stat(dashDir); os.IsNotExist(err) {
				err := os.Mkdir(dashDir, 0755)
				check(err)
			}
		} else {
			total = total + 1
			// get the dashboard json from the grafana api
			db, err := getGrafanaData(grafanaUrl, grafanaToken, fmt.Sprintf("dashboards/uid/%s", ds.UID))
			if err != nil {
				log.Println("Failed to fetch dashboard:", ds.Title)
				failed = failed + 1
				continue
			}
			var dashJSON map[string]interface{}
			err = json.Unmarshal(db, &dashJSON)
			if err != nil {
				log.Println("Failed to parse dashboard:", ds.Title)
				failed = failed + 1
				continue
			}
			// if no folder name is specified for the dashboard, save
			// to the General/ folder
			if ds.FolderTitle == "" {
				ds.FolderTitle = "General"
			}
			dbFolder := fmt.Sprintf("%s/%s", dashDir, ds.FolderTitle)
			if _, err := os.Stat(dbFolder); os.IsNotExist(err) {
				err := os.MkdirAll(dbFolder, 0755)
				check(err)
			}

			dj, err := json.MarshalIndent(dashJSON, "", "  ")
			if err != nil {
				log.Println("Failed to marshal the dashboard JSON:", ds.Title)
				failed = failed + 1
				continue
			}
			err = ioutil.WriteFile(fmt.Sprintf("%s/%s.json", dbFolder, strings.Replace(ds.UID, "/", "-", -1)), dj, 0644)
			if err != nil {
				log.Println("Failed to save dashboard:", ds.Title)
				failed = failed + 1
				continue
			}
			fmt.Println(ds.Title, "downloaded.")
		}
	}
	fmt.Println(fmt.Sprintf("Done! Download Statistics:\n\tTotal: %d\n\tFailed: %d", total, failed))
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("cannot run program")
	}

	grafanaUrl := os.Getenv("GRAFANA_URL")
	if grafanaUrl == "" {
		log.Fatal("Grafana URL is not found")
	}

	grafanaToken := os.Getenv("GRAFANA_TOKEN")
	if grafanaToken == "" {
		log.Fatal("Grafana URL is not found")
	}

	addPaths := os.Getenv("ADD_PATHS")

	backupDir = os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = backupDirDefault
	}

	if os.Args[1] == "backup" {
		endpoint := "search"
		if addPaths != "" {
			endpoint = fmt.Sprintf("search?dashboardUIDs=%s", addPaths)
		}

		resp, err := getGrafanaData(grafanaUrl, grafanaToken, endpoint)
		check(err)

		ds := dashboardSearch{}
		err = json.Unmarshal(resp, &ds)
		check(err)

		syncDashboards(grafanaUrl, grafanaToken, ds)
		return
	}

	err := postDashboards(grafanaUrl, grafanaToken, backupDir)
	check(err)
}

func postDashboards(grafanaUrl, grafanaToken, path string) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if d.IsDir() {
			err := postDashboards(grafanaUrl, grafanaToken, path+"/"+d.Name())
			if err != nil {
				return err
			}
		} else if d.Name() == ".DS_Store" {
			continue
		} else {
			file, err := os.OpenFile(path+"/"+d.Name(), os.O_RDONLY, 0660)
			if err != nil {
				return err
			}
			all, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			var dashWithMetaJSON map[string]any
			err = json.Unmarshal(all, &dashWithMetaJSON)
			if err != nil {
				return err
			}

			dashboardMeta, ok := dashWithMetaJSON["meta"]
			if !ok {
				return errors.New("meta field is not found")
			}
			metaMap := dashboardMeta.(map[string]any)
			folderId := metaMap["folderId"]

			dashboardJson := dashWithMetaJSON["dashboard"].(map[string]any)
			delete(dashboardJson, "id")
			delete(dashboardJson, "uid")
			err = postGrafanaData(grafanaUrl, grafanaToken, "dashboards/db", map[string]any{
				"dashboard": dashboardJson,
				"overwrite": true,
				"folderId":  folderId,
			})
			if err != nil {
				fmt.Printf("Failed to upload %s\n", dashboardJson["title"])
				continue
			}
			fmt.Printf("Dashboard uploaded %s\n", dashboardJson["title"])
		}
	}
	return nil
}

// compress is a function that creates a gzipped
// archive for the specified filepath
func compress(filepath string) (string, error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// the filepath specified does not exist
		return "", err
	}

	now := time.Now()
	// create a timestamped filename.
	// dashboards/ => dashboards-20060102150405.tar.gz
	arcFn := strings.TrimSuffix(filepath, "/") + "-" + now.Format("20060102150405") + ".tar.gz"

	tarGZ := archiver.NewTarGz()
	if err := tarGZ.Archive([]string{filepath}, arcFn); err != nil {
		return "", err
	}

	return arcFn, nil
}
