package endpoint

import (
	"context"

	"github.com/LMF709268224/titan-endpoint-finder/internal/manager"
)

type Client interface {
	GetEndpoint(key string) string
	ResetEndpoints(endpoints map[string][]string)
	GetMyIP() string
}

type clientImpl struct {
	mgr *manager.Manager
}

func NewClient(ctx context.Context, initialEndpoints map[string][]string, serverURL string) Client {
	return &clientImpl{mgr: manager.NewManager(ctx, initialEndpoints, serverURL)}
}

func (c *clientImpl) GetEndpoint(key string) string {
	return c.mgr.SelectOne(key)
}

func (c *clientImpl) ResetEndpoints(endpoints map[string][]string) {
	c.mgr.ResetEndpoints(endpoints)
}

func (c *clientImpl) GetMyIP() string {
	return c.mgr.GetMyIP()
}
