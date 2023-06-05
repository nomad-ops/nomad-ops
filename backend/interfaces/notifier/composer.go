package notifier

import (
	"context"
	"errors"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type ComposerConfig struct {
	Notifiers map[string]application.Notifier
}

// Composer ...
type Composer struct {
	ctx    context.Context
	logger log.Logger
	cfg    ComposerConfig
}

// CreateComposer ...
func CreateComposer(ctx context.Context,
	logger log.Logger,
	cfg ComposerConfig) (*Composer, error) {
	t := &Composer{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
	}

	return t, nil
}

func (s *Composer) Notify(ctx context.Context, opts application.NotifyOptions) error {
	var aggErr error
	for n, notifier := range s.cfg.Notifiers {
		s.logger.LogTrace(ctx, "Notifying %s", n)
		err := notifier.Notify(ctx, opts)
		if err != nil {
			aggErr = errors.Join(aggErr, err)
		}
	}

	return aggErr
}
