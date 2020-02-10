package main

import (
	"flag"
	"fmt"
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/webapp"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"os"
	"strings"
)

var version string = "SNAPSHOT"

type CliConfig struct {
	logLevel      plog.LogLevel
	logTimestamps bool
	logLocations  bool
	printVersion  bool
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUSAGE:\n  %s [OPTIONS] <ZFS_DATASET_NAME>\n\n", os.Args[0])
		fmt.Fprint(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
	}

	cliCfg, zsdCfg := parseFlags()
	setupLogger(cliCfg)

	if cliCfg.printVersion {
		fmt.Printf("zfs-snap-diff: %s\n", version)
		return
	}

	datasetName := flag.Arg(0)
	if len(datasetName) == 0 {
		fmt.Fprintf(os.Stderr, "\nABORT:\n  paramter <ZFS_DATASET_NAME> missing\n")
		if datasetNames, err := zfs.AvailableDatasetNames(zsdCfg.ZFS.UseSudo); err == nil {
			names := strings.Join(datasetNames, " | ")
			fmt.Fprintf(os.Stderr, "\nUSAGE:\n  %s [OPTIONS] <ZFS_DATASET_NAME>\n\n",  os.Args[0])
			fmt.Fprintf(os.Stderr, "  <ZFS_DATASET_NAMES>: %s\n\n", names)
		} else {
			fmt.Fprintf(os.Stderr, "\nUSAGE:\n  %s [OPTIONS] <ZFS_DATASET_NAME>\n\n", os.Args[0])
		}
		fmt.Fprintf(os.Stderr, "For more information use `%s -h`", os.Args[0])
		return
	}

	if z, err := zfs.NewZFS(datasetName, zsdCfg.ZFS); err == nil {
		webapp := webapp.NewWebApp(z, zsdCfg)
		if err := webapp.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "\nUnable to start webapp: %v", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "\nABORT:\n  ")
		fmt.Fprintf(os.Stderr, err.Error())
	}


}

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
	flag.BoolVar((&cliCfg.logLocations), "log-locations", false, "log messages with caller location")

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
	flag.BoolVar(&zfsCfg.MountSnapshots, "mount-snapshots", zfsCfg.MountSnapshots,
		"mount snapshot (only necessary if it's not mounted by zfs automatically")

	flag.Parse()
	return *cliCfg, zsdCfg
}

func setupLogger(cliCfg CliConfig) plog.Logger {

	consoleLogger := plog.NewConsoleLogger(" | ")
	consoleLogger.SetLevel(cliCfg.logLevel)

	if cliCfg.logTimestamps {
		consoleLogger.AddLogFormatter(plog.TimestampUnixDate)
	}

	consoleLogger.AddLogFormatter(plog.Level)

	if cliCfg.logLevel == plog.Trace || cliCfg.logLocations {
		consoleLogger.AddLogFormatter(plog.Location)
	}

	consoleLogger.AddLogFormatter(plog.Message)

	return plog.GlobalLogger().Add(consoleLogger)
}
