package parser

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
)

var patterns = map[string]*regexp.Regexp{
	"vless": regexp.MustCompile(`^vless://([0-9a-f-]+)@([^\s:]+):(\d+)`),
	"vmess": regexp.MustCompile(`^vmess://([A-Za-z0-9+/=]+)`),
	"shadowsocks": regexp.MustCompile(`^ss://([A-Za-z0-9-_+/=]+)@([^:]+):(\d+)`),
	"trojan": regexp.MustCompile(`^trojan://([^\s@]+)@([^\s:]+):(\d+)`),
}

func ValidateURL(protocol, configURL string) bool {
	pattern, exists := patterns[protocol]
	if !exists {
		return false
	}

	if !pattern.MatchString(configURL) {
		return false
	}

	if protocol == "vmess" {
		parts := pattern.FindStringSubmatch(configURL)
		if len(parts) < 2 {
			return false
		}
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			data, err = base64.URLEncoding.DecodeString(parts[1])
			if err != nil {
				return false
			}
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

	if protocol == "shadowsocks" {
		parts := pattern.FindStringSubmatch(configURL)
		if len(parts) < 4 {
			return false
		}
		encoded := parts[1]
		data, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			data, err = base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return false
			}
		}
		creds := strings.SplitN(string(data), ":", 2)
		if len(creds) != 2 {
			return false
		}
		if parts[2] == "" || parts[3] == "" {
			return false
		}
		return true
	}

	return true
}

func ExtractServerAddress(protocol, configURL string) (string, string) {
	pattern, exists := patterns[protocol]
	if !exists {
		return "", ""
	}

	matches := pattern.FindStringSubmatch(configURL)
	if len(matches) == 0 {
		return "", ""
	}

	if protocol == "vmess" {
		if len(matches) < 2 {
			return "", ""
		}
		data, err := base64.StdEncoding.DecodeString(matches[1])
		if err != nil {
			data, err = base64.URLEncoding.DecodeString(matches[1])
			if err != nil {
				return "", ""
			}
		}
		var vmess struct {
			Add string `json:"add"`
			Port string `json:"port"`
		}
		if err := json.Unmarshal(data, &vmess); err != nil {
			return "", ""
		}
		return vmess.Add, vmess.Port
	}

	if protocol == "shadowsocks" {
		if len(matches) < 4 {
			return "", ""
		}
		encoded := matches[1]
		data, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			data, err = base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return "", ""
			}
		}
		if len(matches[2]) == 0 || len(matches[3]) == 0 {
			return "", ""
		}
		return matches[2], matches[3]
	}

	if len(matches) >= 4 {
		return matches[2], matches[3]
	}

	if len(matches) >= 3 {
		return matches[1], matches[2]
	}

	return "", ""
}
