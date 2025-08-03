package parser

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
)

var patterns = map[string]*regexp.Regexp{
	"vless":       regexp.MustCompile(`^vless://([0-9a-f-]+)@([^\s:]+):(\d+).*\?`),
	"vmess":       regexp.MustCompile(`^vmess://([A-Za-z0-9+/=]+)`),
	"shadowsocks": regexp.MustCompile(`^ss://([A-Za-z0-9-_+/=]+)@([^\s:]+):(\d+)`),
	"trojan":      regexp.MustCompile(`^trojan://([^\s@]+)@([^\s:]+):(\d+)`),
}

func ValidateURL(protocol, configURL string) bool {
	if pattern, exists := patterns[protocol]; exists {
		if !pattern.MatchString(configURL) {
			return false
		}
		if protocol == "vmess" {
			// Decode and validate VMess JSON
			parts := patterns["vmess"].FindStringSubmatch(configURL)
			if len(parts) < 2 {
				return false
			}
			data, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return false
			}
			var vmess struct {
				Add string `json:"add"`
				Port string `json:"port"`
			}
			if err := json.Unmarshal(data, &vmess); err != nil {
				return false
			}
			if vmess.Add == "" || vmess.Port == "" {
				return false
			}
			return true
		}
		return true
	}
	return false
}

func ExtractServerAddress(protocol, configURL string) (string, string) {
	if pattern, exists := patterns[protocol]; exists {
		matches := pattern.FindStringSubmatch(configURL)
		if len(matches) >= 3 {
			if protocol == "vmess" {
				data, _ := base64.StdEncoding.DecodeString(matches[1])
				var vmess struct {
					Add string `json:"add"`
					Port string `json:"port"`
				}
				json.Unmarshal(data, &vmess)
				return vmess.Add, vmess.Port
			}
			return matches[2], matches[3]
		}
	}
	return "", ""

}
