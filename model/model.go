package model

// DashboardSearch is a struct that contains
// the search data for the grafana dashboard api
type DashboardSearch []struct {
	ID          int           `json:"id"`
	UID         string        `json:"uid"`
	Title       string        `json:"title"`
	URI         string        `json:"uri"`
	URL         string        `json:"url"`
	Slug        string        `json:"slug"`
	Type        string        `json:"type"`
	Tags        []interface{} `json:"tags"`
	IsStarred   bool          `json:"isStarred"`
	FolderID    int           `json:"folderId,omitempty"`
	FolderUID   string        `json:"folderUid,omitempty"`
	FolderTitle string        `json:"folderTitle,omitempty"`
	FolderURL   string        `json:"folderUrl,omitempty"`
}

type DashboardCreate struct {
	Dashboard any    `json:"dashboard"`
	FolderUid string `json:"folderUid"`
	Overwrite bool   `json:"overwrite"`
}

type BackupInfo struct {
	FailedCount int
	Total       int
}
