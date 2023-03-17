package domain

import (
	"fmt"
	"time"
)

type SourceStatus struct {

	// jobs
	Jobs map[string]JobStatus `json:"jobs,omitempty"`

	// last check time
	// Read Only: true
	LastCheckTime *time.Time `json:"lastCheckTime,omitempty"`

	// last update time
	// Read Only: true
	LastUpdateTime *time.Time `json:"lastUpdateTime,omitempty"`

	// message
	// Read Only: true
	Message string `json:"message,omitempty"`

	// status
	// Read Only: true
	// Enum: [synced error unknown syncing paused init]
	Status string `json:"status,omitempty"`
}

func (s *SourceStatus) DetermineSyncStatus() bool {
	pending := false

	statusMsg := ""
	for key, job := range s.Jobs {
		if job.DeploymentStatus == "running" {
			pending = true
			if statusMsg == "" {
				statusMsg = fmt.Sprintf("Deployment pending for job: %s", key)
			}
		}
		if job.DeploymentStatus == "failed" {
			s.Status = SourceStatusStatusSyncedWithError
			statusMsg = fmt.Sprintf("Deployment failed for job: %s", key)
		}
	}
	if statusMsg != "" {
		s.Message = statusMsg
	}

	return pending
}

const (
	SourceStatusStatusSynced string = "synced"

	SourceStatusStatusSyncedWithError string = "syncedwitherror"

	SourceStatusStatusError string = "error"

	SourceStatusStatusUnknown string = "unknown"

	SourceStatusStatusSyncing string = "syncing"

	SourceStatusStatusPaused string = "paused"

	SourceStatusStatusInit string = "init"
)
