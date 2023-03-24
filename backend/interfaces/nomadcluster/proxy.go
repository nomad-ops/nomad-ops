package nomadcluster

import (
	"context"
	"io"

	"github.com/hashicorp/nomad/api"
)

func (c *Client) ProxyHandler(ctx context.Context, path string, opts api.QueryOptions) (io.ReadCloser, error) {
	c.logger.LogInfo(ctx, "Requesting %s", path)

	resp, err := c.client.Raw().Response(path, &opts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
