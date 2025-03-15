package nomadcluster

import (
	"context"
	"os"
	"testing"

	"github.com/nomad-ops/nomad-ops/backend/application"
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

	// we keep the count at 1
	// but change an env var => it should be an update, BUT the count should remain at 2 (we do not want to undo the work of any autoscaler)
	jobInfo.Job.TaskGroups[0].Tasks[0].Env = map[string]string{
		"TEST": "123",
	}
	info, err = nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !info.Updated {
		t.Fatalf("Job updated")
	}

	runningJob, _, err := nomadClient.client.Jobs().Info("nginx-horizontal-scaling", nomadClient.getQueryOptsCtx(ctx, source, jobInfo))
	if err != nil {
		t.Fatalf("Error getting job info: %v", err)
	}
	if *runningJob.TaskGroups[0].Count != 2 {
		t.Fatalf("Job count should be 2")
	}
}

func Test_IgnoreAutoscalingOptions_Vertical(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test")
		return
	}
	ctx := context.Background()
	logger := log.NewSimpleLogger(true, "Test")

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
	// simulate a vertical scaling by setting the cpu to 600
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

	// we keep the CPU at 500
	// but change an env var => it should be an update, BUT the CPU should remain at 600 (we do not want to undo the work of any autoscaler)
	jobInfo.Job.TaskGroups[0].Tasks[0].Env = map[string]string{
		"TEST": "123",
	}
	info, err = nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !info.Updated {
		t.Fatalf("Job updated")
	}

	runningJob, _, err := nomadClient.client.Jobs().Info("nginx-vertical-scaling", nomadClient.getQueryOptsCtx(ctx, source, jobInfo))
	if err != nil {
		t.Fatalf("Error getting job info: %v", err)
	}
	if *runningJob.TaskGroups[0].Tasks[0].Resources.CPU != 600 {
		t.Fatalf("Task CPU should be 600")
	}
}

func Test_UpdateJob(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test")
		return
	}
	ctx := context.Background()
	logger := log.NewSimpleLogger(true, "Test")

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
	b, err := os.ReadFile("testdata/nginx.hcl")
	if err != nil {
		t.Fatalf("Error reading job file: %v", err)
	}

	jobInfo, err := nomadClient.ParseJob(ctx, string(b))
	if err != nil {
		t.Fatalf("Error parsing job: %v", err)
	}
	jobInfo.GitInfo = application.GitInfo{
		GitCommit: "123456",
	}

	// deploy job
	scaledJob, err := nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !scaledJob.Updated {
		t.Fatalf("Job not updated")
	}

	// deploy job again with different commit
	jobInfo.GitInfo.GitCommit = "654321"
	scaledJob, err = nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if scaledJob.Updated {
		t.Fatalf("Job should NOT have been updated as just the git commit changed")
	}

	// deploy job again with different commit, but with force set
	jobInfo.GitInfo.GitCommit = "654321"
	source.Force = true // forces every commit to be an update
	scaledJob, err = nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !scaledJob.Updated {
		t.Fatalf("Job should have been updated as the git commit changed and force was set")
	}
	source.Force = false

	// deploy job again but force restart
	scaledJob, err = nomadClient.UpdateJob(ctx, source, jobInfo, true)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !scaledJob.Updated {
		t.Fatalf("Job should have been updated as we set the restart flag")
	}

	// deploy job again with different commit, no force, no restart but with a new env
	jobInfo.GitInfo.GitCommit = "789123"
	jobInfo.Job.TaskGroups[0].Tasks[0].Env = map[string]string{
		"TEST": "123",
	}
	scaledJob, err = nomadClient.UpdateJob(ctx, source, jobInfo, false)
	if err != nil {
		t.Fatalf("Error updating job: %v", err)
	}
	if !scaledJob.Updated {
		t.Fatalf("Job should have been updated as the env changed")
	}
}
