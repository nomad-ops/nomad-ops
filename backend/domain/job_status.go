package domain

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
}
