package tester

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
	"github.com/v2root/configpilot/parser"
	"github.com/pires/go-proxyproto"
	"github.com/pingcap/go-ycsb/pkg/util"
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
		return results
	}

	// TLS Handshake
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", server+":"+port, &tls.Config{InsecureSkipVerify: true})
	if err == nil {
		results.TLSOK = true
		conn.Close()
	}

	// ICMP Ping
	pinger, err := util.NewPinger(server)
	if err == nil {
		pinger.Count = 5
		pinger.Timeout = timeout
		pinger.Run()
		stats := pinger.Statistics()
		results.Latency = float64(stats.AvgRtt.Milliseconds())
		results.PacketLoss = float64(stats.PacketLoss)
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
	}

	// DNS Resolution
	start = time.Now()
	_, err = net.LookupHost(server)
	if err == nil {
		results.DNSResolution = float64(time.Since(start).Milliseconds())
	}

	return results
}