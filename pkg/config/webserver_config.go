package config

import (
	"fmt"
)

type WebserverConfig struct {
	// listen
	ListenIp   string `toml:"listen-ip"`
	ListenPort int    `toml:"listen-port"`

	// tls
	UseTLS   bool   `toml:"use-tls"`
	CertFile string `toml:"cert-file"`
	KeyFile  string `toml:"key-file"`
	// webapp
	WebappDir string `toml:"webapp-dir,omitempty"`
}

func (self *WebserverConfig) ListenAddress() string {
	return fmt.Sprintf("%s:%d", self.ListenIp, self.ListenPort)
}

func NewDefaultWebserverConfig() WebserverConfig {
	return WebserverConfig{
		ListenIp:   "127.0.0.1",
		ListenPort: 12345,
		UseTLS:     false,
	}
}
