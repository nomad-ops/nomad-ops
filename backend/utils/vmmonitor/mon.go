package mon

import (
	"context"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/VictoriaMetrics/metrics"

	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type Monitor struct {
	ctx    context.Context
	logger log.Logger
	cfg    Config
	srv    *http.Server
}

type Config struct {
	Address string
}

func StartMon(ctx context.Context, logger log.Logger, cfg Config) (*Monitor, error) {
	m := &Monitor{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
		srv: &http.Server{
			Addr: cfg.Address,
		},
	}

	go func() {
		<-ctx.Done()
		err := m.Stop()
		if err != nil {
			logger.LogError(ctx, "Error stopping server:%v", err)
			return
		}
	}()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, true)
	})

	go func() {
		if err := m.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(ctx, "monitor listen:%+s", err)
		}
	}()

	return m, nil
}

func (m *Monitor) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.srv.Shutdown(ctx)
}
