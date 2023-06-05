package notifier

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"text/template"
	"time"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type WebhookConfig struct {
	WebhookURL          string
	Method              string
	BodyTemplate        string
	QueryParamsTemplate string
	Insecure            bool
	Timeout             time.Duration
	AuthHeaderName      string
	AuthHeaderValue     string
	FireOn              []string
	LogTemplateResults  bool
}

// Webhook ...
type Webhook struct {
	ctx           context.Context
	logger        log.Logger
	cfg           WebhookConfig
	bodyTemplate  *template.Template
	queryTemplate *template.Template
	client        *http.Client
}

// CreateWebhook ...
func CreateWebhook(ctx context.Context,
	logger log.Logger,
	cfg WebhookConfig) (*Webhook, error) {
	if cfg.WebhookURL == "" {
		logger.LogInfo(ctx, "Webhook Webhook URL is empty. Will not notify")
	}
	if cfg.Method == "" && cfg.WebhookURL != "" {
		logger.LogInfo(ctx, "Using the default 'POST' as the webhook method")
		cfg.Method = "POST"
	}

	t := &Webhook{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Insecure},
			},
		},
	}

	fMap := template.FuncMap{
		"queryEscape": func(s string) string {
			return url.QueryEscape(s)
		},
		"now": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"json": func(v interface{}) string {
			b, _ := json.MarshalIndent(v, "", "    ")
			return string(b)
		},
	}

	if cfg.BodyTemplate != "" {
		bodyTmp, err := template.New("body").Funcs(fMap).Parse(cfg.BodyTemplate)
		if err != nil {
			return nil, err
		}
		t.bodyTemplate = bodyTmp
	}
	if cfg.QueryParamsTemplate != "" {
		queryTmp, err := template.New("query").Funcs(fMap).Parse(cfg.QueryParamsTemplate)
		if err != nil {
			return nil, err
		}
		t.queryTemplate = queryTmp
	}

	return t, nil
}

func (s *Webhook) Notify(ctx context.Context, opts application.NotifyOptions) error {
	if s.cfg.WebhookURL == "" {
		return nil
	}
	fire := false
	for _, v := range s.cfg.FireOn {
		if v == string(opts.Type) {
			fire = true
		}
	}
	if !fire && len(s.cfg.FireOn) != 0 {
		return nil
	}

	var r io.Reader
	if s.bodyTemplate != nil {
		// apply a body template
		b := &bytes.Buffer{}
		err := s.bodyTemplate.Execute(b, opts)
		if err != nil {
			return err
		}
		if s.cfg.LogTemplateResults {
			s.logger.LogInfo(ctx, "%s %s:\n%s", s.cfg.Method, s.cfg.WebhookURL, b.String())
		}
		r = b
	}

	req, err := http.NewRequestWithContext(ctx, s.cfg.Method, s.cfg.WebhookURL, r)
	if err != nil {
		return err
	}

	if s.queryTemplate != nil {
		// apply a query template
		b := &bytes.Buffer{}
		err := s.queryTemplate.Execute(b, opts)
		if err != nil {
			return err
		}
		values, err := url.ParseQuery(b.String())
		if err != nil {
			return err
		}
		req.URL.RawQuery = values.Encode()
	}

	req.Header.Set("Content-Type", "application/json")

	if s.cfg.LogTemplateResults {
		reqB, _ := httputil.DumpRequestOut(req, true)
		s.logger.LogInfo(ctx, "%s %s:\n%s", s.cfg.Method, s.cfg.WebhookURL, string(reqB))
	}
	if s.cfg.AuthHeaderName != "" && s.cfg.AuthHeaderValue != "" {
		req.Header.Set(s.cfg.AuthHeaderName, s.cfg.AuthHeaderValue)
	}

	resp, err := http.DefaultClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		reqB, _ := httputil.DumpRequestOut(req, true)
		respB, _ := httputil.DumpResponse(resp, true)
		s.logger.LogError(ctx, "Could not send Webhook Message:%v - %v", string(reqB), string(respB))
		return fmt.Errorf("could not send Webhook Message")
	}

	return nil
}
