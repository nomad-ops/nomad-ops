package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/nomad/api"

	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

var (
	ErrNotFound = errors.New("errNotFound")
)

type ClusterState struct {
	CurrentJobs map[string]*JobInfo
}
type DesiredState struct {
	GitInfo GitInfo
	Jobs    map[string]*JobInfo
}

type GitInfo struct {
	GitCommit string
}

type JobInfo struct {
	GitInfo GitInfo
	*api.Job
}

type JobParser interface {
	ParseJob(ctx context.Context, j string) (*JobInfo, error)
}

type GetCurrentClusterStateOptions struct {
	Source *domain.Source
}

type UpdateJobInfo struct {
	Updated          bool
	Created          bool
	DeploymentStatus DeploymentStatus
}

type DeploymentStatus struct {
	Status string
}

type ClusterAPI interface {
	GetCurrentClusterState(ctx context.Context, opts GetCurrentClusterStateOptions) (*ClusterState, error)
	UpdateJob(ctx context.Context, src *domain.Source, job *JobInfo, restart bool) (*UpdateJobInfo, error)
	DeleteJob(ctx context.Context, src *domain.Source, job *JobInfo) error
}

type ChangeInfo struct {
	DryRun bool
	Create map[string]*JobInfo
	Delete map[string]*JobInfo
	Update map[string]*JobInfo
}

type ReconcilerFunc func(ctx context.Context,
	src *domain.Source,
	desiredState *DesiredState,
	restart bool) (*domain.SourceStatus, *ChangeInfo, error)

func (r *ReconciliationManager) OnReconcile(ctx context.Context,
	src *domain.Source,
	desiredState *DesiredState,
	restart bool) (*domain.SourceStatus, *ChangeInfo, error) {

	currentState, err := r.clusterAccess.GetCurrentClusterState(ctx, GetCurrentClusterStateOptions{
		Source: src,
	})
	if err != nil {
		return nil, nil, err
	}

	changed := &ChangeInfo{
		DryRun: src.Paused,
		Create: map[string]*JobInfo{},
		Delete: map[string]*JobInfo{},
		Update: map[string]*JobInfo{},
	}

	res := &domain.SourceStatus{
		Jobs:          map[string]domain.JobStatus{},
		Status:        domain.SourceStatusStatusSynced,
		LastCheckTime: toTimePtr(time.Now()),
	}

	for k, job := range currentState.CurrentJobs {
		if _, ok := desiredState.Jobs[k]; !ok {
			r.logger.LogTrace(ctx, "Checking if job is still required: %v...%+v", strPtrToStr(job.Name), log.ToJSONString(job))
			cpy := job

			if cpy.Periodic != nil && cpy.Periodic.Enabled != nil && *cpy.Periodic.Enabled {
				// ignore periodic jobs
				continue
			}
			if cpy.ParentID == nil || *cpy.ParentID == "" {
				// has a parent job, periodic probably
				continue
			}

			changed.Delete[k] = cpy

			if src.Paused {
				r.logger.LogInfo(ctx, "Found job %s that is no longer desired. Would be deleted...", k)
				continue
			}

			r.logger.LogInfo(ctx, "Found job %s that is no longer desired. Deleting...", k)
			err := r.clusterAccess.DeleteJob(ctx, src, job)
			if err != nil {
				return nil, nil, err
			}

			// we have a change
			res.LastUpdateTime = toTimePtr(time.Now())

			ev := &domain.Event{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Deleted Job:%v", strPtrToStr(job.Job.Name)),
				Type:      domain.EventTypeDeleted,
				Source:    src,
			}
			err = r.evRepo.SaveEvent(ctx, ev)
			if err != nil {
				r.logger.LogError(ctx, "Could not store event:%v", log.ToJSONString(ev))
			}
			r.logger.LogInfo(ctx, "Found job %s that is no longer desired. Deleting...Done", k)
		}
	}

	for k, job := range desiredState.Jobs {
		r.logger.LogTrace(ctx, "Updating job %v...%+v", strPtrToStr(job.Name), log.ToJSONString(job))
		info, err := r.clusterAccess.UpdateJob(ctx, src, job, restart)
		if err != nil {
			r.logger.LogInfo(ctx, "Could not UpdateJob %v", log.ToJSONString(job))
			return nil, nil, err
		}

		jobStatus := domain.JobStatus{
			Type:             strPtrToStr(job.Type),
			Status:           "unknown",
			DeploymentStatus: info.DeploymentStatus.Status,
			Groups:           map[string]domain.GroupStatus{},
		}
		if j, ok := currentState.CurrentJobs[k]; ok {
			jobStatus.Status = strPtrToStr(j.Status)
			jobStatus.StatusDescription = strPtrToStr(j.StatusDescription)
		}
		for _, tg := range job.TaskGroups {
			groupStatus := domain.GroupStatus{
				Count:    intPtrToInt(tg.Count),
				Services: map[string]domain.ServiceStatus{},
				Tasks:    map[string]domain.TaskStatus{},
			}
			for _, t := range tg.Tasks {
				taskStatus := domain.TaskStatus{
					Driver: t.Driver,
				}
				groupStatus.Tasks[t.Name] = taskStatus
			}
			for _, svc := range tg.Services {
				svcStatus := domain.ServiceStatus{
					Port: svc.PortLabel,
				}
				groupStatus.Services[svc.Name] = svcStatus
			}
			jobStatus.Groups[strPtrToStr(tg.Name)] = groupStatus
		}

		res.Jobs[strPtrToStr(job.Name)] = jobStatus

		r.logger.LogTrace(ctx, "Updating job %v...Done", strPtrToStr(job.Name))

		if !info.Created && !info.Updated {
			r.logger.LogTrace(ctx, "Nothing to do for job %v", strPtrToStr(job.Name))
			continue
		}

		// we have a change
		res.LastUpdateTime = toTimePtr(time.Now())

		if info.Created {
			cpy := job
			changed.Create[k] = cpy

			if src.Paused {
				r.logger.LogInfo(ctx, "Would create job %v", strPtrToStr(job.Name))
				continue
			}
			ev := &domain.Event{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Created Job:%v", strPtrToStr(job.Job.Name)),
				Type:      domain.EventTypeCreated,
				Source:    src,
			}
			err = r.evRepo.SaveEvent(ctx, ev)
			if err != nil {
				r.logger.LogError(ctx, "Could not store event:%v", log.ToJSONString(ev))
			}
			r.logger.LogInfo(ctx, "Created job %v", strPtrToStr(job.Name))
		}
		if info.Updated {
			cpy := job
			changed.Update[k] = cpy

			if src.Paused {
				r.logger.LogInfo(ctx, "Would update job %v", strPtrToStr(job.Name))
				continue
			}

			ev := &domain.Event{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("Updated Job:%v", strPtrToStr(job.Job.Name)),
				Type:      domain.EventTypeUpdated,
				Source:    src,
			}
			err = r.evRepo.SaveEvent(ctx, ev)
			if err != nil {
				r.logger.LogError(ctx, "Could not store event:%v - %v", err, log.ToJSONString(ev))
			}
			r.logger.LogInfo(ctx, "Updated job %v", strPtrToStr(job.Name))
			err = r.notifier.Notify(ctx, NotifyOptions{
				Source:  src,
				GitInfo: desiredState.GitInfo,
				Type:    NotificationSuccess,
				Message: fmt.Sprintf("Updated Job:%v", strPtrToStr(job.Job.Name)),
				Infos: []NotifyAdditionalInfos{
					{
						Header: "Git-Commit",
						Text:   desiredState.GitInfo.GitCommit,
					},
					{
						Header: "Git-Url",
						Text:   src.URL,
					},
					{
						Header: "Git-Ref",
						Text:   src.Branch,
					},
					{
						Header: "Git-Repo-Path",
						Text:   src.Path,
					},
					{
						Header: "Nomad-Namespace",
						Text:   src.Namespace,
					},
					{
						Header: "Nomad-Region",
						Text:   src.Region,
					},
					{
						Header: "Force Restart",
						Text:   fmt.Sprintf("%v", restart),
					},
				},
			})
			if err != nil {
				r.logger.LogError(ctx, "Could not notify:%v", err)
			}
		}
	}

	return res, changed, nil
}

func toTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func strPtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intPtrToInt(s *int) int {
	if s == nil {
		return 0
	}
	return *s
}
