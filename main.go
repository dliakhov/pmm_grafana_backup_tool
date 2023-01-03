package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dliakhov/pmm/grafana/backup/cmd"
	"github.com/dliakhov/pmm/grafana/backup/services"
)

type Config struct {
	GrafanaURL   string
	GrafanaToken string
	AddPaths     string
	BackupDir    string
}

type applicationContext struct {
	backupHandler  *cmd.BackupHandler
	restoreHandler *cmd.RestoreHandler
	grafanaService services.GrafanaService
	storeDashboard services.StoreDashboard
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("cannot run program")
	}

	config := parseConfig()
	context := initServices(config)

	switch os.Args[1] {
	case "backup":
		err := context.backupHandler.Handle()
		if err != nil {
			log.Fatal(err)
		}
	case "restore":
		err := context.restoreHandler.Handle()
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Unknown command")
	}
}

func initServices(config Config) applicationContext {
	httpClient := http.DefaultClient
	grafanaService := services.NewGrafanaService(httpClient, config.GrafanaURL, config.GrafanaToken)
	storeDashboard := services.NewStoreDashboards(config.BackupDir, grafanaService)

	backupHandler := cmd.NewBackupHandler(grafanaService, storeDashboard)
	restoreHandler := cmd.NewRestoreHandler(storeDashboard)
	return applicationContext{
		grafanaService: grafanaService,
		backupHandler:  backupHandler,
		storeDashboard: storeDashboard,
		restoreHandler: restoreHandler,
	}
}

func parseConfig() Config {
	config := Config{}
	grafanaUrl := os.Getenv("GRAFANA_URL")
	if grafanaUrl == "" {
		log.Fatal("Grafana URL is not found")
	}
	config.GrafanaURL = grafanaUrl

	grafanaToken := os.Getenv("GRAFANA_TOKEN")
	if grafanaToken == "" {
		log.Fatal("Grafana URL is not found")
	}
	config.GrafanaToken = grafanaToken

	addPaths := os.Getenv("ADD_PATHS")
	config.AddPaths = addPaths

	backupDir := os.Getenv("BACKUP_DIR")
	config.BackupDir = backupDir

	return config
}
