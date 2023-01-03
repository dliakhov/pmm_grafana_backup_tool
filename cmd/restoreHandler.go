package cmd

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/dliakhov/pmm/grafana/backup/services"
)

type RestoreHandler struct {
	storeDashboard services.StoreDashboard
}

func NewRestoreHandler(storeDashboard services.StoreDashboard) *RestoreHandler {
	return &RestoreHandler{
		storeDashboard: storeDashboard,
	}
}

func (b *RestoreHandler) Handle() error {
	err := b.storeDashboard.RestoreDashboards()
	if err != nil {
		return errors.Wrap(err, "error happened during restoring")
	}
	fmt.Println("Restored all dashboards successfully")
	return nil
}
