package notifier

import (
	"context"
	"fmt"
	"testing"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

func TestWebhook(t *testing.T) {
	ctx := context.Background()
	logger := log.NewSimpleLogger(false, "Test")
	s, err := CreateWebhook(ctx, logger, WebhookConfig{
		WebhookURL:         "",
		Method:             "POST",
		LogTemplateResults: true,
		FireOn: []string{
			"success",
		},
		BodyTemplate: `{
	"timestamp": "{{ now }}",
	"project": "my project",
	"category": "infrastructure",
	"link": "some link",
	"version": "{{ .GitInfo.GitCommit }}",
	"agent": "whatever",
	"description": "{{ .Message }}"
}`,
	})
	if err != nil {
		t.Errorf("Could not CreateSlack:%v", err)
		return
	}
	err = s.Notify(ctx, application.NotifyOptions{
		Source: &domain.Source{
			ID:   "testid",
			Name: "polygons-stage",
		},
		GitInfo: application.GitInfo{
			GitCommit: "5dc8ecf",
		},
		Type:    application.NotificationError,
		Message: "Updated polygons",
		Infos: []application.NotifyAdditionalInfos{
			{
				Header: "Git-Url",
				Text:   "https://github.com/trivago/polygons",
			},
			{
				Header: "Git-Rev",
				Text:   "main",
			},
			{
				Header: "Git-Repo-Path",
				Text:   "/deployments",
			},
			{
				Header: "Nomad-Namespace",
				Text:   "beta-test",
			},
			{
				Header: "Nomad-Region",
				Text:   "",
			},
			{
				Header: "Force Restart",
				Text:   fmt.Sprintf("%v", true),
			},
			{
				Header: "Error",
				Text:   fmt.Sprintf("Could not Reconcile:%v", fmt.Errorf("something went wrong. really really really really really really really really really really bad")),
				Large:  true,
			},
		},
	})
	if err != nil {
		t.Errorf("Could not Notify:%v", err)
		return
	}
}
