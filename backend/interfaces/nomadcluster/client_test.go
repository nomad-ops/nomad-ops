package nomadcluster

import (
	"context"
	"os"
	"testing"

	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

// need a running nomad cluster to test this
// set NOMAD_ADDR and NOMAD_TOKEN env vars if needed
func Test_IgnoreAutoscalingOptions_Horizontal(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test")
		return
	}
	ctx := context.Background()
	logger := log.NewSimpleLogger(true, "Test")

	//os.Setenv("NOMAD_ADDR", "http://172.31.101.5:4646")

	nomadClient, err := CreateClient(ctx, logger, ClientConfig{
		NomadToken: "",
	})
	if err != nil {
		t.Fatalf("Error creating nomad client: %v", err)
	}

	// prepare test data
	source := &domain.Source{
		ID:              "test",
		Name:            "test",
		Branch:          "main",
		Namespace:       "test",
		CreateNamespace: true,
	}
	b, err := os.ReadFile("testdata/nginx_horizontal_scaling.hcl")
	if err != nil {
		t.Fatalf("Error reading job file: %v", err)
	}

	jobInfo, err := nomadClient.ParseJob(ctx, string(b))
	if err != nil {
		t.Fatalf("Error parsing job: %v", err)
	}
	// simulate a scaling by setting the count to 2
	count := 2
	jobInfo.Job.TaskGroups[0].Count = &count

	// Prepare environment

	// deploy job
	scaledJob, err := nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !scaledJob.Updated {
		t.Fatalf("Job not updated")
	}

	// we now have a job with count 2 running
	// now we will update the job with count 1 => it should be a no-op
	count = 1
	jobInfo.Job.TaskGroups[0].Count = &count
	info, err := nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if info.Updated {
		t.Fatalf("Job updated")
	}
}

func Test_IgnoreAutoscalingOptions_Vertical(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test")
		return
	}
	ctx := context.Background()
	logger := log.NewSimpleLogger(true, "Test")

	os.Setenv("NOMAD_ADDR", "http://172.31.101.5:4646")

	nomadClient, err := CreateClient(ctx, logger, ClientConfig{
		NomadToken: "",
	})
	if err != nil {
		t.Fatalf("Error creating nomad client: %v", err)
	}

	// prepare test data
	source := &domain.Source{
		ID:              "test",
		Name:            "test",
		Branch:          "main",
		Namespace:       "test",
		CreateNamespace: true,
	}
	b, err := os.ReadFile("testdata/nginx_vertical_scaling.hcl")
	if err != nil {
		t.Fatalf("Error reading job file: %v", err)
	}

	jobInfo, err := nomadClient.ParseJob(ctx, string(b))
	if err != nil {
		t.Fatalf("Error parsing job: %v", err)
	}
	// simulate a vertical scaling by setting the cpu to 6000
	cpu := 600
	jobInfo.Job.TaskGroups[0].Tasks[0].Resources.CPU = &cpu

	// Prepare environment

	// deploy job
	scaledJob, err := nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !scaledJob.Updated {
		t.Fatalf("Job not updated")
	}

	// we now have a task with 600 CPU running
	// now we will update the job with CPU 500 => it should be a no-op
	cpu = 500
	jobInfo.Job.TaskGroups[0].Tasks[0].Resources.CPU = &cpu
	info, err := nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if info.Updated {
		t.Fatalf("Job updated")
	}
}
