package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/dliakhov/pmm/grafana/backup/model"
)

type StoreDashboard interface {
	SaveDashboards(dashboards model.DashboardSearch) (model.BackupInfo, error)
	RestoreDashboards() error
}

type storeDashboardsFileSystemImpl struct {
	backupDir      string
	grafanaService GrafanaService
}

func NewStoreDashboards(backupDir string, service GrafanaService) *storeDashboardsFileSystemImpl {
	return &storeDashboardsFileSystemImpl{
		backupDir:      backupDir,
		grafanaService: service,
	}
}

func (s *storeDashboardsFileSystemImpl) SaveDashboards(dashboards model.DashboardSearch) (model.BackupInfo, error) {
	if _, err := os.Stat(s.backupDir); !os.IsNotExist(err) {
		os.RemoveAll(s.backupDir)
	}
	//err := os.Mkdir(dashDir, 0755)
	//check(err)

	var failed, total int
	log.Println("Syncing Dashboards...")
	for _, ds := range dashboards {
		if ds.Type == "dash-folder" {
			// check if a folder for exists, if not, create one
			if _, err := os.Stat(s.backupDir); os.IsNotExist(err) {
				err := os.Mkdir(s.backupDir, 0755)
				if err != nil {
					return model.BackupInfo{}, errors.Wrap(err, "cannot create backup directory")
				}
			}
		} else {
			total = total + 1
			// get the dashboard json from the grafana api
			dashJSON, err := s.grafanaService.GetDashboard(ds.UID)
			if err != nil {
				log.Println("Failed to fetch dashboard:", ds.Title)
				failed = failed + 1
				continue
			}
			// if no folder name is specified for the dashboard, save
			// to the General/ folder
			if ds.FolderTitle == "" {
				ds.FolderTitle = "General"
			}
			dbFolder := fmt.Sprintf("%s/%s", s.backupDir, ds.FolderTitle)
			if _, err := os.Stat(dbFolder); os.IsNotExist(err) {
				err := os.MkdirAll(dbFolder, 0755)
				if err != nil {
					return model.BackupInfo{}, errors.Wrap(err, "cannot create sub directory")
				}
			}

			dj, err := json.MarshalIndent(dashJSON, "", "  ")
			if err != nil {
				log.Println("Failed to marshal the dashboard JSON:", ds.Title)
				failed = failed + 1
				continue
			}
			err = os.WriteFile(fmt.Sprintf("%s/%s.json", dbFolder, strings.Replace(ds.UID, "/", "-", -1)), dj, 0644)
			if err != nil {
				log.Println("Failed to save dashboard:", ds.Title)
				failed = failed + 1
				continue
			}
			fmt.Println(ds.Title, "downloaded.")
		}
	}
	return model.BackupInfo{
		FailedCount: failed,
		Total:       total,
	}, nil
}

func (s *storeDashboardsFileSystemImpl) RestoreDashboards() error {
	return s.restoreDashboards(s.backupDir)
}

func (s *storeDashboardsFileSystemImpl) restoreDashboards(path string) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if d.IsDir() {
			err := s.restoreDashboards(path + "/" + d.Name())
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
			err = s.grafanaService.CreateOrUpdateDashboards(map[string]any{
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
