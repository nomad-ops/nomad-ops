package domain

type GroupStatus struct {

	// count
	Count int `json:"count,omitempty"`

	// services
	Services map[string]ServiceStatus `json:"services,omitempty"`

	// tasks
	Tasks map[string]TaskStatus `json:"tasks,omitempty"`
}
