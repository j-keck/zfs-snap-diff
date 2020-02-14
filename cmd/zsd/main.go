package main

import (
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"github.com/j-keck/zfs-snap-diff/pkg/scanner"
	"github.com/j-keck/plog"
	"os"
	"path/filepath"
	"fmt"
	"time"
	"flag"
	"math"
	"strings"
	"strconv"
	"errors"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
)

var version string = "SNAPSHOT"

type CliConfig struct {
	logLevel     plog.LogLevel
	printVersion bool
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "zsd is a little cli tool to restore a file from a zfs-snapshot.\n\n")
		fmt.Fprintf(os.Stderr, "USAGE:\n %s [OPTIONS] <FILE> <ACTION>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nACTIONS:\n")
		fmt.Fprintf(os.Stderr, "  list                : list zfs-snapshots with different file-versions for the given file\n")
		fmt.Fprintf(os.Stderr, "  diff    <#|SNAPSHOT>: show differences between the actual version and the selected version\n")
		fmt.Fprintf(os.Stderr, "  restore <#|SNAPSHOT>: restore the file to the given version\n")
		fmt.Fprintf(os.Stderr, "\nzsd is a part of zfs-snap-diff (https://j-keck.github.io/zfs-snap-diff)\n")
	}


	cliCfg := parseFlags()

	if cliCfg.printVersion {
		fmt.Printf("zsd: %s\n", version)
		return
	}

	log := setupLogger(cliCfg)

	if len(flag.Args()) < 2 {
		fmt.Fprintf(os.Stderr, "Argument <FILE> <ACTION> missing (check %s -h)\n", os.Args[0])
		return
	}


	// file path
	fileName := flag.Arg(0)
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Errorf("unable to get absolute path for: '%s' - %v", fileName, err)
		return
	}


	// init zfs handler
	zfs, ds, err := zfs.NewZFSForFilePath(filePath)
	if err != nil {
		log.Errorf("unable to get zfs handler for path: '%s' - %v", filePath, err)
		return
	}
	log.Debugf("work on dataset: %s", ds.Name)


    // action
	action := flag.Arg(1)
	switch action {
	case "list":
		fmt.Printf("scan the last %d days for other file versions\n", config.Get.DaysToScan)
		dr := scanner.NewDateRange(time.Now(), config.Get.DaysToScan)
		sc := scanner.NewScanner(dr, "auto", ds, zfs)
		scanResult, err := sc.FindFileVersions(filePath)
		if err != nil {
			log.Errorf("scan failed - %v", err)
			return
		}

		cacheFileVersions(scanResult.FileVersions)

		// find the longest snapshot name
		width := 0
		for _, v := range scanResult.FileVersions {
			width = int(math.Max(float64(width), float64(len(v.Snapshot.Name))))

		}

		// show snapshots where the file was modified
		header := fmt.Sprintf("%3s | %-[2]*s | %s", "#", width, "Snapshot", "Snapshot age")
		fmt.Printf("%s\n%s\n", header, strings.Repeat("-", len(header)))
		for idx, v := range scanResult.FileVersions {
			age := humanDuration(time.Since(v.Snapshot.Created))
			fmt.Printf("%3d | %-[2]*s | %s\n", idx, width, v.Snapshot.Name, age)
		}

	case "diff":
		if len(flag.Args()) != 3 {
			fmt.Fprintf(os.Stderr, "Argument <#|SNAPSHOT> missing (check %s -h)\n", os.Args[0])
			return
		}

		version, err := lookupRequestedVersion(flag.Arg(2))
		if err != nil {
			log.Error(err)
			return
		}

		diffs, err := diff.NewDiffFromPath(version.Backup.Path, filePath, 5)
		if err != nil {
			log.Errorf("unable to create diff - %v", err)
			return
		}

		fmt.Printf("Diff from the actual version to the version from: %s\n", version.Backup.MTime)
		fmt.Printf("%s", diffs.PrettyTextDiff)

	case "restore":
		if len(flag.Args()) != 3 {
			fmt.Fprintf(os.Stderr, "Argument <#|SNAPSHOT> missing (check %s -h)\n", os.Args[0])
			return
		}

		version, err := lookupRequestedVersion(flag.Arg(2))
		if err != nil {
			log.Error(err)
			return
		}


		backupPath, err := version.Actual.Backup()
		if err != nil {
			log.Errorf("unable to backup the acutal version - %v", err)
			return
		}
		fmt.Printf("backup from the actual version created at: %s\n", backupPath)

		// restore the backup version
		version.Backup.Copy(version.Actual.Path)
		fmt.Printf("version restored from snapshot: %s\n", version.Snapshot.Name)

	default:
		fmt.Fprintf(os.Stderr, "invalid action: %s\n", action)
		return
	}
}


func lookupRequestedVersion(arg string) (*scanner.FileVersion, error) {

	// load file-versions from cache file
	fileVersions, err := loadCachedFileVersions()
	if err != nil {
		return nil, err
	}


	if idx, err := strconv.Atoi(arg); err == nil {
		if idx >= 0 && idx < len(fileVersions) {
			return &fileVersions[idx], nil
		}
		return nil, errors.New("invalid version index given")
	} else {
		for _, v := range fileVersions {
			if v.Snapshot.Name == arg {
				return &v, nil
			}
		}
	}


	return nil, errors.New("requested version not found")
}


func humanDuration(dur time.Duration) string {
	s := int(dur.Seconds())
	if s < 60 {
		return fmt.Sprintf("%d seconds", s)
	}

	m := int(dur.Minutes())
	if m < 60 {
		return fmt.Sprintf("%d minutes", m)
	}
	h := int(dur.Hours())
	if h < 24 {
		return fmt.Sprintf("%d hours", h)
	}

	d := int(h / 24)
	return fmt.Sprintf("%d days", d)
}

func parseFlags() CliConfig {
	loadConfig()

	cliCfg := new(CliConfig)

	// cli
	flag.BoolVar(&cliCfg.printVersion, "V", false, "print version and exit")
	flag.IntVar(&config.Get.DaysToScan, "d", config.Get.DaysToScan, "days to scan")

	// logging
	cliCfg.logLevel = plog.Note
	plog.FlagDebugVar(&cliCfg.logLevel, "v", "debug output")
	plog.FlagTraceVar(&cliCfg.logLevel, "vv", "trace output with caller location")

	flag.Parse()
	return *cliCfg
}

func loadConfig() {
	plog.DropUnhandledMessages()
	configDir, _ := fs.ConfigDir()
	configPath := configDir.Path + "/zfs-snap-diff.toml"
	config.LoadConfig(configPath)
}


func setupLogger(cliCfg CliConfig) plog.Logger {

	log := plog.NewConsoleLogger(" ")
	log.SetLevel(cliCfg.logLevel)
	log.AddLogFormatter(plog.LevelFmt("%5s: "))

	if cliCfg.logLevel == plog.Trace {
		log.AddLogFormatter(plog.Location)
	}

	log.AddLogFormatter(plog.Message)

	return plog.GlobalLogger().Add(log)
}
