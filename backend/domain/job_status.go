package domain

import "encoding/json"

type JobStatus struct {

	// groups
	Groups map[string]GroupStatus `json:"groups,omitempty"`

	// status
	Status string `json:"status,omitempty"`

	// deploymentStatus
	// pending | ok | failed
	DeploymentStatus string `json:"deploymentStatus,omitempty"`

	// status description
	StatusDescription string `json:"statusDescription,omitempty"`

	// type
	Type string `json:"type,omitempty"`

	// namespace
	Namespace string `json:"namespace,omitempty"`

	// diff
	Diff json.RawMessage `json:"diff,omitempty"`
}
