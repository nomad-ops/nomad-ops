package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/nomad-ops/nomad-ops/backend/application"
	"github.com/nomad-ops/nomad-ops/backend/utils/log"
)

type SlackConfig struct {
	WebhookURL  string
	BaseURL     string
	IconSuccess string
	IconError   string
	EnvInfoText string
}

// Slack ...
type Slack struct {
	ctx    context.Context
	logger log.Logger
	cfg    SlackConfig
}

// CreateSlack ...
func CreateSlack(ctx context.Context,
	logger log.Logger,
	cfg SlackConfig) (*Slack, error) {
	if cfg.WebhookURL == "" {
		logger.LogInfo(ctx, "Slack Webhook URL is empty. Will not notify")
	}
	t := &Slack{
		ctx:    ctx,
		logger: logger,
		cfg:    cfg,
	}

	return t, nil
}

type messageRequest struct {
	Blocks []Block `json:"blocks,omitempty"`
}
type Text struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}
type Accessory struct {
	Type     string         `json:"type,omitempty"`
	Text     *AccessoryText `json:"text,omitempty"`
	Value    string         `json:"value,omitempty"`
	URL      string         `json:"url,omitempty"`
	ActionID string         `json:"action_id,omitempty"`
}
type AccessoryText struct {
	Type  string `json:"type,omitempty"`
	Text  string `json:"text,omitempty"`
	Emoji bool   `json:"emoji,omitempty"`
}
type Block struct {
	Type      string     `json:"type,omitempty"`
	Text      *Text      `json:"text,omitempty"`
	Fields    []Text     `json:"fields,omitempty"`
	Accessory *Accessory `json:"accessory,omitempty"`
}

func (s *Slack) Notify(ctx context.Context, opts application.NotifyOptions) error {
	if s.cfg.WebhookURL == "" {
		return nil
	}

	icon := s.cfg.IconSuccess
	if opts.Type == application.NotificationError {
		icon = s.cfg.IconError
	}
	icon = icon + " "

	msg := messageRequest{
		Blocks: []Block{
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: icon + opts.Message,
				},
			},
		},
	}

	currBlock := Block{
		Type: "section",
	}

	for _, i := range opts.Infos {
		if i.Large {
			if len(currBlock.Fields) > 0 {
				msg.Blocks = append(msg.Blocks, currBlock)
				currBlock = Block{
					Type: "section",
				}
			}

			msg.Blocks = append(msg.Blocks, Block{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*%s:*\n%s", i.Header, i.Text),
				},
			})
			continue
		}
		currBlock.Fields = append(currBlock.Fields, Text{
			Type: "mrkdwn",
			Text: fmt.Sprintf("*%s:*\n%s", i.Header, i.Text),
		})
	}
	if len(currBlock.Fields) > 0 {
		msg.Blocks = append(msg.Blocks, currBlock)
	}

	if opts.Source != nil {
		msg.Blocks = append(msg.Blocks, Block{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: fmt.Sprintf("<%s|View at Nomad Ops>", s.cfg.BaseURL+opts.Source.ID),
			},
		})
	}

	msg.Blocks = append(msg.Blocks, Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: s.cfg.EnvInfoText,
		},
	})

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Post(s.cfg.WebhookURL, "application/json", bytes.NewBuffer(b))
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		respB, _ := httputil.DumpResponse(resp, true)
		s.logger.LogError(ctx, "Could not send Webhook Message:%v - %v", string(b), string(respB))
		return fmt.Errorf("could not send Webhook Message")
	}

	return nil
}
