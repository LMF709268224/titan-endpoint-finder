package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	fetchInterval = 30 * time.Minute
	checkInterval = 10 * time.Minute
)

type Endpoint struct {
	Address    string
	Status     bool
	muEndpoint sync.RWMutex
}

type Manager struct {
	muMap        sync.RWMutex
	endpointsMap map[string][]*Endpoint

	fetchEndpointURL string
	fetchCancel      context.CancelFunc

	myIP string
}

func NewManager(ctx context.Context, initialEndpoints map[string][]string, serverURL string) *Manager {
	m := &Manager{fetchEndpointURL: serverURL}
	m.ResetEndpoints(initialEndpoints)

	m.checkEndpoints()

	go m.startHealthCheck(ctx)

	m.startFetchingEndpoints(ctx)

	return m
}

func (m *Manager) GetMyIP() string {
	return m.myIP
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
	defer m.muMap.RUnlock()

	endpoints, ok := m.endpointsMap[key]
	if !ok {
		return ""
	}

	for _, ep := range endpoints {
		// ep := m.endpoints[i]

		ep.muEndpoint.RLock()
		status := ep.Status
		ep.muEndpoint.RUnlock()
		if status {
			return ep.Address
		}
	}

	return ""
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

	if m.fetchEndpointURL == "" {
		return
	}

	fetchCtx, cancel := context.WithCancel(ctx)
	m.fetchCancel = cancel

	go m.fetchEndpoints(fetchCtx)
}

func (m *Manager) fetchEndpoints(ctx context.Context) {
	ticker := time.NewTicker(fetchInterval)
	defer ticker.Stop()

	m.doFetchEndpoints()

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func (m *Manager) checkHealth(addr string) error {
	addr = "http://" + addr

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	m.myIP = string(respBody)
	return nil
}
