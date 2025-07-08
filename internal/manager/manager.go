package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	fetchInterval = 30 * time.Minute
	checkInterval = 10 * time.Minute

	connectTimeout = 15 * time.Second
)

// Endpoint 接入点
type Endpoint struct {
	Address    string
	Status     bool
	muEndpoint sync.RWMutex
}

// Manager Endpoint Manager
type Manager struct {
	muMap        sync.RWMutex
	endpointsMap map[string][]*Endpoint

	fetchEndpointURL string
	fetchCancel      context.CancelFunc

	clientPublicIP string
}

// NewManager create
func NewManager(ctx context.Context, initialEndpoints map[string][]string, serverURL string) (*Manager, error) {
	m := &Manager{fetchEndpointURL: serverURL}
	m.ResetEndpoints(initialEndpoints)

	err := m.doFetchEndpoints()
	if err != nil {
		return nil, err
	}

	m.checkEndpoints()

	go m.startHealthCheck(ctx)
	m.startFetchingEndpoints(ctx)

	return m, nil
}

// GetClientPublicIP 获取客户端源IP
func (m *Manager) GetClientPublicIP() string {
	return m.clientPublicIP
}

func (m *Manager) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			m.checkEndpoints()
		}
	}
}

// SelectOne 获取可用接入点
func (m *Manager) SelectOne(key string) string {
	m.muMap.RLock()
	endpoints, ok := m.endpointsMap[key]
	m.muMap.RUnlock()
	if !ok || len(endpoints) == 0 {
		return ""
	}

	// 收集所有可用节点
	var available []*Endpoint
	for _, ep := range endpoints {
		ep.muEndpoint.RLock()
		if ep.Status {
			available = append(available, ep)
		}
		ep.muEndpoint.RUnlock()
	}

	if len(available) == 0 {
		return ""
	}

	// 随机选择可用节点
	selected := available[rand.Intn(len(available))]
	return selected.Address
}

// 检查所有接入点状态
func (m *Manager) checkEndpoints() {
	m.muMap.RLock()
	endpointsMap := m.endpointsMap
	m.muMap.RUnlock()

	var wg sync.WaitGroup
	for _, endpointsList := range endpointsMap {
		for _, ep := range endpointsList {
			wg.Add(1)
			go func(ep *Endpoint) {
				defer wg.Done()
				err := m.checkHealth(ep.Address)
				if err != nil {
					log.Printf("checkHealth %s failed: %s", ep.Address, err.Error())
				}

				ep.muEndpoint.Lock()
				ep.Status = err == nil
				ep.muEndpoint.Unlock()
			}(ep)
		}
	}
	wg.Wait()
}

// ResetEndpoints 重置接入点
func (m *Manager) ResetEndpoints(initialEndpoints map[string][]string) {
	m.muMap.Lock()
	defer m.muMap.Unlock()

	newEndpoints := make(map[string][]*Endpoint)
	for key, addrs := range initialEndpoints {
		for _, addr := range addrs {
			newEndpoints[key] = append(newEndpoints[key], &Endpoint{
				Address: addr,
				Status:  true,
			})
		}
	}
	m.endpointsMap = newEndpoints
}

func (m *Manager) startFetchingEndpoints(ctx context.Context) {
	if m.fetchCancel != nil {
		m.fetchCancel()
		m.fetchCancel = nil
	}

	fetchCtx, cancel := context.WithCancel(ctx)
	m.fetchCancel = cancel

	go m.fetchEndpoints(fetchCtx)
}

func (m *Manager) fetchEndpoints(ctx context.Context) {
	ticker := time.NewTicker(fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := m.doFetchEndpoints()
			if err != nil {
				log.Printf("doFetchEndpoints %s failed: %s", m.fetchEndpointURL, err.Error())
			}
		}
	}
}

func (m *Manager) doFetchEndpoints() error {
	if m.fetchEndpointURL == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", m.fetchEndpointURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status: %d", resp.StatusCode)
	}

	var newEndpoints map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&newEndpoints); err != nil {
		return err
	}

	m.ResetEndpoints(newEndpoints)
	return nil
}
