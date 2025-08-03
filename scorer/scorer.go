package scorer

import (
	"github.com/v2root/configpilot/tester"
)

func CalculateScore(results tester.TestResults) float64 {
	score := 0.0
	if results.TLSOK {
		score += 25
	}
	if results.Latency > 0 && results.Latency < 50 {
		score += 30
	} else if results.Latency < 100 {
		score += 15
	}
	if results.Throughput > 5 {
		score += 30
	} else if results.Throughput > 1 {
		score += 10
	}
	if results.DNSResolution > 0 && results.DNSResolution < 100 {
		score += 10
	}
	if results.PacketLoss == 0 {
		score += 5
	}
	return score
}