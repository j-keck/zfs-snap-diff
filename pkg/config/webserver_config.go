package config

import (
	"fmt"
)

type WebserverConfig struct {
	// listen
	ListenIp              string
	ListenPort            int
	ListenOnAllInterfaces bool
	// tls
	UseTLS   bool
	CertFile string
	KeyFile  string
	// webapp
	WebappDir string
}

func (self *WebserverConfig) ListenAddress() string {
	if self.ListenOnAllInterfaces {
		if self.ListenIp != "127.0.0.1" {
			log.Warnf("ignore 'ListenIp' value: '%s' because of 'ListenOnAllInterfaces' was set",
				self.ListenIp)
		}
		return fmt.Sprintf("0.0.0.0:%d", self.ListenPort)
	}
	return fmt.Sprintf("%s:%d", self.ListenIp, self.ListenPort)
}

func NewDefaultWebserverConfig() WebserverConfig {
	return WebserverConfig{
		ListenIp:              "127.0.0.1",
		ListenPort:            12345,
		ListenOnAllInterfaces: false,
		UseTLS:                false,
	}
}
