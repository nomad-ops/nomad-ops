package nomadcluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// types "github.com/hashicorp/nomad-openapi/clients/go/v1"
	// v1 "github.com/hashicorp/nomad-openapi/v1"

	"github.com/hashicorp/nomad/api"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

var (
	metaKeyOps          = "nomadops"
	metaKeySrcID        = "nomadopssrcid"
	metaKeySrcUrl       = "nomadopssrcurl"
	metaKeySrcCommit    = "nomadopssrccommit"
	metaKeyForceRestart = "nomadopsforcerestart"
)

type ClientConfig struct {
	NomadToken string
}

type Client struct {
	ctx    context.Context
	logger log.Logger
	cfg    ClientConfig
	client *api.Client
	url    string
}

func CreateClient(ctx context.Context,
	logger log.Logger,
	cfg ClientConfig) (*Client, error) {

	defCfg := api.DefaultConfig()

	if cfg.NomadToken != "" {
		// Use default client config from ENV, optionally a custom token
		defCfg.SecretID = cfg.NomadToken
	}

	client, err := api.NewClient(defCfg)

	if err != nil {
		return nil, err
	}

	c := &Client{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
		client: client,
		url:    defCfg.Address,
	}

	return c, nil
}

func (c *Client) SubscribeJobChanges(ctx context.Context, cb func(jobName string)) error {
	var index uint64 = 0
	if _, meta, err := c.client.Jobs().List(nil); err == nil {
		index = meta.LastIndex
	}

	queryOptions := &api.QueryOptions{
		Namespace: "*",
	}

	eventCh, err := c.client.EventStream().Stream(ctx, map[api.Topic][]string{
		api.TopicJob:        {"*"},
		api.TopicDeployment: {"*"},
	}, index, queryOptions.WithContext(ctx))
	if err != nil {
		return err
	}

	eventHandler := func(event *api.Events) {
		for _, e := range event.Events {

			c.logger.LogTrace(ctx, "Received nomad event:%v", e.Type)

			switch e.Type {
			case "JobRegistered", "JobDeregistered":

				job, err := e.Job()
				if err != nil {
					return
				}
				if job == nil || job.ID == nil {
					c.logger.LogInfo(ctx, "Received no Job on '%s': %s", e.Type, log.ToJSONString(e))
					return
				}

				cb(*job.ID)
			case "DeploymentStatusUpdate":
				dep, err := e.Deployment()
				if err != nil {
					return
				}
				if dep == nil {
					c.logger.LogInfo(ctx, "Received no deployment on 'DeploymentStatusUpdate': %s", log.ToJSONString(e))
					return
				}
				cb(dep.JobID)
			default:
			}
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case events := <-eventCh:

				if events.IsHeartbeat() {
					continue
				}

				eventHandler(events)
			}
		}
	}()

	return nil
}

func hasUpdate(ctx context.Context,
	logger log.Logger,
	job *application.JobInfo,
	diffResp *api.JobPlanResponse,
	restart,
	force bool) bool {
	logger.LogTrace(ctx, "hasUpdate() - Job:%v", log.ToJSONString(job))

	hasGroupScaling := func(taskGroupName string) bool {
		for _, taskGroup := range job.Job.TaskGroups {
			if *taskGroup.Name != taskGroupName {
				continue
			}
			if taskGroup.Scaling != nil && taskGroup.Scaling.Enabled != nil && *taskGroup.Scaling.Enabled {
				return true
			}
		}
		return false
	}

	hasTaskScaling := func(taskGroupName, taskName, resourceName string) bool {
		for _, taskGroup := range job.Job.TaskGroups {
			if *taskGroup.Name != taskGroupName {
				continue
			}
			for _, task := range taskGroup.Tasks {
				if task.Name != taskName {
					continue
				}
				for _, res := range task.ScalingPolicies {
					if res.Type != resourceName {
						continue
					}
					return true
				}
			}
		}
		return false
	}

	hasDiff := false
	if len(diffResp.Diff.Objects) > 0 {
		return true
	}
	fieldDiff := diffResp.Diff.Fields
	if len(fieldDiff) > 0 {
		// if only the git commit change we will not see it as a change
		// if only the forced restart is a change we will not see it as a change either
		// use force to update it anyway
		if len(fieldDiff) != 1 ||
			(fieldDiff[0].Name != fmt.Sprintf("Meta[%s]", metaKeySrcCommit) &&
				fieldDiff[0].Name != fmt.Sprintf("Meta[%s]", metaKeyForceRestart)) ||
			force || restart {
			return true
		}
	}
	for _, taskGrp := range diffResp.Diff.TaskGroups {
		if len(taskGrp.Fields) > 0 {
			// Check if only count is changed && scaling is enabled
			if len(taskGrp.Fields) > 1 || taskGrp.Fields[0].Name != "Count" || !hasGroupScaling(taskGrp.Name) {
				return true
			}
		}
		if len(taskGrp.Objects) > 0 {
			// Check if only Scaling is changed
			// Check if scaling options are changed
			if len(taskGrp.Objects) > 1 || taskGrp.Objects[0].Name != "Scaling" || len(taskGrp.Objects[0].Fields) > 0 {
				return true
			}
		}

		for _, task := range taskGrp.Tasks {
			if len(task.Fields) > 0 {
				return true
			}
			if len(task.Objects) > 0 {
				if len(task.Objects) > 1 {
					return true
				}
				if task.Objects[0].Name != "Resources" {
					return true
				}
				if !hasTaskScaling(taskGrp.Name, task.Name, "vertical_cpu") &&
					!hasTaskScaling(taskGrp.Name, task.Name, "vertical_mem") {
					// no scaling for this task, update it
					return true
				}

				onlyCPUAndMemoryChanged := true
				cpuChanged := false
				memoryChanged := false

				for _, resourceEdit := range task.Objects[0].Fields {
					if resourceEdit.Name != "CPU" && resourceEdit.Name != "MemoryMB" && resourceEdit.Type != "None" {
						// something aside from CPU and MemoryMB is changed
						onlyCPUAndMemoryChanged = false
					}
					if resourceEdit.Name == "CPU" {
						cpuChanged = true
					}
					if resourceEdit.Name == "MemoryMB" {
						memoryChanged = true
					}
				}

				if !onlyCPUAndMemoryChanged {
					return true
				}

				if cpuChanged && !hasTaskScaling(taskGrp.Name, task.Name, "vertical_cpu") {
					return true
				}
				if memoryChanged && !hasTaskScaling(taskGrp.Name, task.Name, "vertical_mem") {
					return true
				}
			}
		}
	}
	return hasDiff
}

func (c *Client) ParseJob(ctx context.Context, j string) (*application.JobInfo, error) {
	parsedJob, err := c.client.Jobs().ParseHCL(j, false)
	if err != nil {
		return nil, err
	}

	return &application.JobInfo{
		Job: parsedJob,
	}, nil
}

func (c *Client) getQueryOptsCtx(ctx context.Context, src *domain.Source, job *application.JobInfo) *api.QueryOptions {

	opts := &api.QueryOptions{}
	if job != nil && job.Namespace != nil && *job.Namespace != "" {
		opts.Namespace = *job.Namespace
	}
	if job != nil && job.Region != nil && *job.Region != "" {
		opts.Region = *job.Region
	}

	// Src overrides job
	if src.Namespace != "" {
		opts.Namespace = src.Namespace
	}
	if src.Region != "" {
		opts.Region = src.Region
	}

	return opts.WithContext(ctx)
}

func (c *Client) getWriteOptions(ctx context.Context, src *domain.Source, job *application.JobInfo) *api.WriteOptions {

	opts := &api.WriteOptions{}
	if job != nil && job.Namespace != nil && *job.Namespace != "" {
		opts.Namespace = *job.Namespace
	}
	if job != nil && job.Region != nil && *job.Region != "" {
		opts.Region = *job.Region
	}

	// Src overrides job
	if src.Namespace != "" {
		opts.Namespace = src.Namespace
	}
	if src.Region != "" {
		opts.Region = src.Region
	}

	return opts.WithContext(ctx)
}

func (c *Client) UpdateJob(ctx context.Context,
	src *domain.Source,
	job *application.JobInfo,
	restart bool) (*application.UpdateJobInfo, error) {

	if src.CreateNamespace {
		writeOptions := c.getWriteOptions(ctx, src, job)
		if writeOptions.Namespace == "" {
			return nil, fmt.Errorf("require a namespace to be set in conjunction with 'CreateNamespace'")
		}
		// Make sure that namespace exists
		_, err := c.client.Namespaces().Register(&api.Namespace{
			Name: writeOptions.Namespace,
			Meta: map[string]string{
				metaKeyOps: "true",
			},
		}, c.getWriteOptions(ctx, src, job))
		if err != nil {
			return nil, err
		}
	}

	metadata := job.Job.Meta
	if metadata == nil {
		metadata = map[string]string{}
	}

	// claiming this job as our job!
	metadata[metaKeyOps] = "true"
	metadata[metaKeySrcUrl] = src.URL
	metadata[metaKeySrcID] = src.ID
	metadata[metaKeySrcCommit] = job.GitInfo.GitCommit

	if restart {
		metadata[metaKeyForceRestart] = time.Now().Format(time.RFC3339Nano)
	}

	job.Meta = metadata
	resp, _, err := c.client.Jobs().Plan(job.Job, true, c.getWriteOptions(ctx, src, job))

	if err != nil {
		return nil, err
	}

	deploymentStatus := ""

	deployment, _, err := c.client.Jobs().LatestDeployment(*job.ID, c.getQueryOptsCtx(ctx, src, job))
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "not found") {
			// low effort "not found" detection
			return nil, err
		}
	}
	if deployment != nil {
		deploymentStatus = deployment.Status
		c.logger.LogTrace(ctx, "DeploymentStatus:%s %v", *job.ID, deploymentStatus)
	}

	if !hasUpdate(ctx, c.logger, job, resp, restart, src.Force) {
		c.logger.LogTrace(ctx, "Job is already up to date.")

		return &application.UpdateJobInfo{
			DeploymentStatus: application.DeploymentStatus{
				Status: deploymentStatus,
			},
		}, nil
	}

	c.logger.LogTrace(ctx, "Job Diff:%v", log.ToJSONString(resp.Diff))

	if !src.Paused {
		regResp, _, err := c.client.Jobs().Register(job.Job, c.getWriteOptions(ctx, src, job))
		if err != nil {
			return nil, err
		}

		c.logger.LogInfo(ctx, "Job Post:%v", log.ToJSONString(regResp))
	}

	return &application.UpdateJobInfo{
		Updated: true, // TODO check for creation, for now everything is an update...which is kinda true
		Diff:    json.RawMessage(log.ToJSONString(resp.Diff)),
		DeploymentStatus: application.DeploymentStatus{
			Status: deploymentStatus,
		},
	}, nil
}

func (c *Client) DeleteJob(ctx context.Context, src *domain.Source, job *application.JobInfo) error {

	_, _, err := c.client.Jobs().Deregister(*job.Job.Name, false, c.getWriteOptions(ctx, src, job))

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetURL(ctx context.Context) (string, error) {
	return c.url, nil
}

func (c *Client) GetCurrentClusterState(ctx context.Context,
	opts application.GetCurrentClusterStateOptions) (*application.ClusterState, error) {

	queryOptions := &api.QueryOptions{
		Namespace: "*", // Query all authorized namespaces
		Params: map[string]string{
			"meta": "true",
		},
		Filter: fmt.Sprintf(`"nomadopssrcid" in Meta and Meta["nomadopssrcid"] == "%s"`, opts.Source.ID),
	}
	joblist, _, err := c.client.Jobs().List(queryOptions.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	clusterState := &application.ClusterState{
		CurrentJobs: map[string]*application.JobInfo{},
	}

	for _, job := range joblist {
		m := job.Meta
		// Ignore stuff that is not managed by us
		if len(m) == 0 {
			continue
		}
		// only consider jobs with my source id!
		if m[metaKeySrcID] != opts.Source.ID {
			continue
		}

		queryOptions := &api.QueryOptions{
			Namespace: job.Namespace,
		}

		j, _, err := c.client.Jobs().Info(job.Name, queryOptions.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		clusterState.CurrentJobs[job.Name] = &application.JobInfo{
			Job: j,
		}
	}

	return clusterState, nil
}
