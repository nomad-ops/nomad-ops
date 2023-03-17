package notifier

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

func TestSlack(t *testing.T) {
	ctx := context.Background()
	logger := log.NewSimpleLogger(false, "Test")
	s, err := CreateSlack(ctx, logger, SlackConfig{
		WebhookURL:  os.Getenv("TEST_SLACK_WEBHOOK"),
		BaseURL:     "https://nomad-ops.prod.eu.tcs.trv.cloud/ui/sources/",
		IconSuccess: ":check:",
		IconError:   ":check-no:",
	})
	if err != nil {
		t.Errorf("Could not CreateSlack:%v", err)
		return
	}
	err = s.Notify(ctx, application.NotifyOptions{
		Source: &domain.Source{
			ID: "testid",
		},
		Type:    application.NotificationError,
		Message: "Could not Reconcile",
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
