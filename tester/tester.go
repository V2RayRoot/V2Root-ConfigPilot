package tester

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
	"github.com/go-ping/ping"
	"github.com/v2root/configpilot/parser"
	"golang.org/x/exp/slog"
)

type TestResults struct {
	TLSOK          bool
	Latency        float64 // ms
	Throughput     float64 // Mbps
	DNSResolution  float64 // ms
	PacketLoss     float64 // %
}

func TestConfig(protocol, configURL string, timeout time.Duration) TestResults {
	results := TestResults{}
	server, port := parser.ExtractServerAddress(protocol, configURL)
	if server == "" || port == "" {
		slog.Warn("Invalid server or port", "protocol", protocol, "url", configURL)
		return results
	}

	// TLS Handshake
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", server+":"+port, &tls.Config{InsecureSkipVerify: true})
	if err == nil {
		results.TLSOK = true
		conn.Close()
	} else {
		slog.Warn("TLS handshake failed", "server", server, "error", err)
	}

	// ICMP Ping with TCP fallback
	pinger, err := ping.NewPinger(server)
	if err == nil {
		pinger.Count = 5
		pinger.Timeout = timeout
		pinger.SetPrivileged(false) // Non-privileged mode for GitHub Actions
		err = pinger.Run()
		if err == nil {
			stats := pinger.Statistics()
			results.Latency = float64(stats.AvgRtt.Milliseconds())
			results.PacketLoss = float64(stats.PacketLoss)
		} else {
			slog.Warn("ICMP ping failed, falling back to TCP", "server", server, "error", err)
			// TCP fallback
			start := time.Now()
			conn, err := net.DialTimeout("tcp", server+":"+port, timeout)
			if err == nil {
				results.Latency = float64(time.Since(start).Milliseconds())
				conn.Close()
			} else {
				slog.Warn("TCP ping fallback failed", "server", server, "error", err)
			}
		}
	} else {
		slog.Warn("Failed to create pinger, falling back to TCP", "server", server, "error", err)
		// TCP fallback
		start := time.Now()
		conn, err := net.DialTimeout("tcp", server+":"+port, timeout)
		if err == nil {
			results.Latency = float64(time.Since(start).Milliseconds())
			conn.Close()
		} else {
			slog.Warn("TCP ping fallback failed", "server", server, "error", err)
		}
	}

	// Throughput Test
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Timeout: timeout}).DialContext,
		},
	}
	start := time.Now()
	resp, err := client.Get("https://" + server + ":" + port + "/1mb.bin")
	if err == nil {
		defer resp.Body.Close()
		duration := time.Since(start).Seconds()
		results.Throughput = (1 * 8) / duration // 1 MiB in Mbps
	} else {
		slog.Warn("Throughput test failed", "server", server, "error", err)
	}

	// DNS Resolution
	start = time.Now()
	_, err = net.LookupHost(server)
	if err == nil {
		results.DNSResolution = float64(time.Since(start).Milliseconds())
	} else {
		slog.Warn("DNS resolution failed", "server", server, "error", err)
	}

	return results
}
