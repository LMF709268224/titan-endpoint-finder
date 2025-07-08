package manager

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func (m *Manager) checkHealth(addr string) error {
	err := m.checkTCPHealth(addr)
	if err != nil {
		return err
	}

	return m.checkUDPHealth(addr)
}

func (m *Manager) checkTCPHealth(addr string) error {
	addr = "http://" + addr

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return fmt.Errorf("TCP new request err: %w", err)
	}
	client := &http.Client{Timeout: connectTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("TCP connect err: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("TCP read err: %w", err)
	}

	m.clientPublicIP = string(respBody)
	return nil
}

func (m *Manager) checkUDPHealth(addr string) error {
	// addr = "udp://" + addr

	conn, err := net.DialTimeout("udp", addr, connectTimeout)
	if err != nil {
		return fmt.Errorf("UDP connect err: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(connectTimeout)
	conn.SetDeadline(deadline)

	probe := []byte("HEALTH_CHECK\x00")
	if _, err := conn.Write(probe); err != nil {
		return fmt.Errorf("UDP conn write err: %w", err)
	}

	return nil
}
