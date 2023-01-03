package cmd

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/dliakhov/pmm/grafana/backup/services"
)

type BackupHandler struct {
	grafanaService services.GrafanaService
	storeDashboard services.StoreDashboard
}

func NewBackupHandler(grafanaService services.GrafanaService, storeDashboard services.StoreDashboard) *BackupHandler {
	return &BackupHandler{
		grafanaService: grafanaService,
		storeDashboard: storeDashboard,
	}
}

func (b *BackupHandler) Handle() error {
	dashboards, err := b.grafanaService.GetAllDashboards()
	if err != nil {
		return errors.Wrap(err, "cannot backup the dashboards")
	}

	backupInfo, err := b.storeDashboard.SaveDashboards(dashboards)
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Done! Download Statistics:\n\tTotal: %d\n\tFailed: %d", backupInfo.Total, backupInfo.FailedCount))
	return nil
}
