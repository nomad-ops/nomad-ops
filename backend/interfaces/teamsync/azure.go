package teamsync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"github.com/nomad-ops/nomad-ops/backend/domain"
	"github.com/nomad-ops/nomad-ops/backend/utils/errors"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type AzureTeamSync struct {
	ctx    context.Context
	logger log.Logger
	cfg    AzureTeamSyncConfig
}

type AzureTeamSyncConfig struct {
	TeamNameProperty string
}

func CreateAzureTeamSync(ctx context.Context,
	logger log.Logger,
	cfg AzureTeamSyncConfig) (*AzureTeamSync, error) {
	t := &AzureTeamSync{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
	}

	return t, nil
}

func (s *AzureTeamSync) GetTeam(ctx context.Context, e *core.RecordAuthWithOAuth2Event) (*domain.Team, error) {
	if s.cfg.TeamNameProperty == "" {
		// No property set => always return "not found"
		return nil, errors.ErrNotFound
	}
	req, err := http.NewRequestWithContext(ctx,
		"GET",
		fmt.Sprintf("%s%s?$select=%s",
			`https://graph.microsoft.com/v1.0/users/`,
			e.OAuth2User.Id,
			s.cfg.TeamNameProperty), nil)
	if err != nil {
		s.logger.LogError(ctx, "Could not get userinfo:%v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", e.OAuth2User.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.LogError(ctx, "Could not get userinfo:%v", err)
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.LogError(ctx, "Could not get userinfo (readall):%v", err)
		return nil, err
	}

	d := map[string]string{}
	err = json.Unmarshal(b, &d)
	if err != nil {
		s.logger.LogError(ctx, "Could not get userinfo (unmarshal):%v", err)
		return nil, err
	}
	if d[s.cfg.TeamNameProperty] == "" {
		return nil, errors.ErrNotFound
	}
	return &domain.Team{
		Name: d[s.cfg.TeamNameProperty],
	}, nil
}
