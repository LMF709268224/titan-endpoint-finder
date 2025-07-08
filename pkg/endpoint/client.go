package endpoint

import (
	"context"

	"github.com/LMF709268224/titan-endpoint-finder/internal/manager"
)

// Client client
type Client interface {
	GetEndpoint(key string) string
	ResetEndpoints(endpoints map[string][]string)
	GetClientPublicIP() string
}

type clientImpl struct {
	mgr *manager.Manager
}

// NewClient create
func NewClient(ctx context.Context, initialEndpoints map[string][]string, serverURL string) (Client, error) {
	mgr, err := manager.NewManager(ctx, initialEndpoints, serverURL)
	if err != nil {
		return nil, err
	}
	return &clientImpl{mgr: mgr}, nil
}

func (c *clientImpl) GetEndpoint(key string) string {
	return c.mgr.SelectOne(key)
}

func (c *clientImpl) ResetEndpoints(endpoints map[string][]string) {
	c.mgr.ResetEndpoints(endpoints)
}

func (c *clientImpl) GetClientPublicIP() string {
	return c.mgr.GetClientPublicIP()
}
