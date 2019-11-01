package main

import (
	"flag"
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
)

type CliConfig struct {
	logLevel      plog.LogLevel
	logTimestamps bool
	printVersion  bool
}

// this file contains code copies from the new 'zfs-snap-diff' binary
func parseFlags() (CliConfig, config.Config) {
	cliCfg := new(CliConfig)
	zsdCfg := config.NewDefaultConfig()

	// cli
	flag.BoolVar(&cliCfg.printVersion, "V", false, "print version and exit")

	// logging
	cliCfg.logLevel = plog.Info
	plog.FlagDebugVar(&cliCfg.logLevel, "v", "debug output")
	plog.FlagTraceVar(&cliCfg.logLevel, "vv", "trace output with caller location")
	flag.BoolVar((&cliCfg.logTimestamps), "log-timestamps", false, "log messages with timestamps in unix format")

	// webserver
	webCfg := &zsdCfg.Webserver
	flag.StringVar(&webCfg.ListenIp, "l", webCfg.ListenIp, "webserver listen address")
	flag.IntVar(&webCfg.ListenPort, "p", webCfg.ListenPort, "webserver port")
	flag.BoolVar(&webCfg.ListenOnAllInterfaces, "a", webCfg.ListenOnAllInterfaces, "listen on all interfaces")
	flag.BoolVar(&webCfg.UseTLS, "tls", webCfg.UseTLS,
		"use TLS - NOTE: -cert <CERT_FILE> -key <KEY_FILE> are mandatory")
	flag.StringVar(&webCfg.CertFile, "cert", webCfg.CertFile, "TLS certificate file")
	flag.StringVar(&webCfg.KeyFile, "key", webCfg.KeyFile, "TLS private key file")
	flag.StringVar(&webCfg.WebappDir, "webapp-dir", webCfg.WebappDir,
		"when given, serve the webapp from the given directory")

	// zfs
	zfsCfg := &zsdCfg.ZFS
	flag.BoolVar(&zfsCfg.UseSudo, "use-sudo", zfsCfg.UseSudo, "use sudo when executing 'zfs' commands")
	flag.BoolVar(&zfsCfg.MountSnapshot, "mount-snapshot", zfsCfg.MountSnapshot,
		"mount snapshot (only necessary if it's not mounted by zfs automatically")

	flag.Parse()
	return *cliCfg, zsdCfg
}

func setupLogger(cliCfg CliConfig) {

	consoleLogger := plog.NewConsoleLogger(" | ")
	consoleLogger.SetLevel(cliCfg.logLevel)

	if cliCfg.logTimestamps {
		consoleLogger.AddLogFormatter(plog.TimestampUnixDate)
	}

	consoleLogger.AddLogFormatter(plog.Level)

	if cliCfg.logLevel == plog.Trace {
		consoleLogger.AddLogFormatter(plog.Location)
	}

	consoleLogger.AddLogFormatter(plog.Message)
	plog.GlobalLogger().Add(consoleLogger)
}
