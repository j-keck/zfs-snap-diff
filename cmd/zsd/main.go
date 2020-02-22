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
	zsdBin := os.Args[0]
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "zsd - cli tool to find older versions of a given file in your zfs snapshots.\n\n")
		fmt.Fprintf(os.Stderr, "USAGE:\n %s [OPTIONS] <FILE> <ACTION>\n\n", zsdBin)
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nACTIONS:\n")
		fmt.Fprintf(os.Stderr, "  list                : list zfs snapshots where the given file was modified\n")
		fmt.Fprintf(os.Stderr, "  cat     <#|SNAPSHOT>: show the file content from the given snapshot\n");
		fmt.Fprintf(os.Stderr, "  diff    <#|SNAPSHOT>: show a diff from the selected snapshot to the actual version\n")
		fmt.Fprintf(os.Stderr, "  restore <#|SNAPSHOT>: restore the file from the given snapshot\n")
		fmt.Fprintf(os.Stderr, "\nYou can use the snapshot number from the `list` output or the snapshot name to select a snapshot.\n")
		fmt.Fprintf(os.Stderr, "\nProject home page: https://j-keck.github.io/zfs-snap-diff\n")
	}


	initLogger()
	cliCfg := parseFlags()
	log := reconfigureLogger(cliCfg)

	if cliCfg.printVersion {
		fmt.Printf("zsd: %s\n", version)
		return
	}

	if len(flag.Args()) < 2 {
		fmt.Fprintf(os.Stderr, "Argument <FILE> <ACTION> missing (see `%s -h` for help)\n", zsdBin)
		return
	}


	// file path
	fileName := flag.Arg(0)
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Errorf("unable to get absolute path for: '%s' - %v", fileName, err)
		return
	}
	log.Debugf("full path: %s", filePath)


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

		// find the longest snapshot name to format the output table
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

	case "cat":
		if len(flag.Args()) != 3 {
			fmt.Fprintf(os.Stderr, "Argument <#|SNAPSHOT> missing (see `%s -h` for help)\n", zsdBin)
			return
		}

		versionName := flag.Arg(2)
		version, err := lookupRequestedVersion(filePath, versionName)
		if err != nil {
			log.Error(err)
			return
		}

		file, err := fs.GetFileHandle(version.Backup.Path)
		if err !=nil {
			log.Errorf("unable to find file in the snapshot - %v", err)
			return
		}

		content, err := file.ReadString()
		if err != nil {
			log.Errorf("unable to get content from %s - %v", file.Name, err)
			return
		}

		fmt.Println(content)

	case "diff":
		if len(flag.Args()) != 3 {
			fmt.Fprintf(os.Stderr, "Argument <#|SNAPSHOT> missing (see `%s -h` for help)\n", zsdBin)
			return
		}

		versionName := flag.Arg(2)
		version, err := lookupRequestedVersion(filePath, versionName)
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
			fmt.Fprintf(os.Stderr, "Argument <#|SNAPSHOT> missing (see `%s -h` for help)\n", zsdBin)
			return
		}

		versionName := flag.Arg(2)
		version, err := lookupRequestedVersion(filePath, versionName)
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
		fmt.Fprintf(os.Stderr, "invalid action: %s (see `%s -h` for help)\n", action, zsdBin)
		return
	}
}


func lookupRequestedVersion(filePath, versionName string) (*scanner.FileVersion, error) {

	// load file-versions from cache file
	fileVersions, err := loadCachedFileVersions()
	if err != nil {
		return nil, err
	}


	// `versionName` can be the snapshot number from the `list` output or the name
	var version *scanner.FileVersion
	if idx, err := strconv.Atoi(versionName); err == nil {
		if idx >= 0 && idx < len(fileVersions) {
			version = &fileVersions[idx]
		} else {
			return nil, errors.New("snapshot number not found")
		}
	} else {
		for _, v := range fileVersions {
			if v.Snapshot.Name == versionName {
				version = &v
				break
			}
		}
		if version == nil {
			return nil, errors.New("snapshot name not found")
		}
	}

	if version.Actual.Path == filePath {
		return version, nil
	} else {
		return nil, errors.New("file mismatch - perform a `list` action at first")
	}
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

	// zfs
	zfsCfg := &config.Get.ZFS
	flag.BoolVar(&zfsCfg.UseSudo, "use-sudo", zfsCfg.UseSudo, "use sudo when executing 'zfs' commands")
	flag.BoolVar(&zfsCfg.MountSnapshots, "mount-snapshots", zfsCfg.MountSnapshots,
		"mount snapshot (only necessary if it's not mounted by zfs automatically")

	flag.Parse()
	return *cliCfg
}

func loadConfig() {
	plog.DropUnhandledMessages()
	configDir, _ := fs.ConfigDir()
	configPath := configDir.Path + "/zfs-snap-diff.toml"
	config.LoadConfig(configPath)
}

func initLogger() {
	consoleLogger := plog.NewConsoleLogger(" | ")
	consoleLogger.AddLogFormatter(plog.Level)
	consoleLogger.AddLogFormatter(plog.Message)

	plog.GlobalLogger().Add(consoleLogger)
}

func reconfigureLogger(cliCfg CliConfig) plog.Logger {

	consoleLogger := plog.NewConsoleLogger(" | ")
	consoleLogger.SetLevel(cliCfg.logLevel)
	consoleLogger.AddLogFormatter(plog.Level)

	if cliCfg.logLevel == plog.Trace {
		consoleLogger.AddLogFormatter(plog.Location)
	}

	consoleLogger.AddLogFormatter(plog.Message)

	return plog.GlobalLogger().Reset().Add(consoleLogger)
}
