package main

import (
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"github.com/j-keck/zfs-snap-diff/pkg/diff"
	"github.com/j-keck/zfs-snap-diff/pkg/config"
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
	"github.com/j-keck/zfs-snap-diff/pkg/fs"
)

var version string = "SNAPSHOT"

type CliConfig struct {
	logLevel     plog.LogLevel
	printVersion bool
	scanDays     int
}
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUSAGE:\n %s [OPTIONS] <FILE> <ACTION>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nACTIONS:\n")
		fmt.Fprintf(os.Stderr, "  list: list snapshots with different file-versions for the given file\n")
		fmt.Fprintf(os.Stderr, "  diff <#|SNAPSHOT>: show differences\n")
		//fmt.Fprintf(os.Stderr, "  restore <#|SNAPSHOT>: restore the file to the given version\n")
	}

	cliCfg, zfsCfg := parseFlags()

	if cliCfg.printVersion {
		fmt.Printf("zsd: %s\n", version)
		return
	}

	log := setupLogger(cliCfg)

	if len(flag.Args()) < 2 {
		fmt.Fprintf(os.Stderr, "Arguments missing - use '%s -h'\n", os.Args[0])
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
	zfs, ds, err := zfs.NewZFSForFilePath(filePath, zfsCfg)
	if err != nil {
		log.Errorf("unable to get zfs handler for path: '%s' - %v", filePath, err)
		return
	}
	log.Debugf("work on dataset: %s", ds.Name)


    // action
	action := flag.Arg(1)
	switch action {
	case "list":
		log.Debugf("scan the last %d days for other file versions", cliCfg.scanDays)
		dr := scanner.NewDateRange(time.Now(), cliCfg.scanDays)
		sc := scanner.NewScanner(dr, "auto", ds, zfs)
		scanResult, err := sc.FindFileVersions(filePath)
		if err != nil {
			log.Errorf("scan failed - %v", err)
			return
		}

		width := 0
		for _, v := range scanResult.FileVersions {
			width = int(math.Max(float64(width), float64(len(v.Snapshot.Name))))

		}

		header := fmt.Sprintf("%3s | %-[2]*s | %s", "#", width, "Snapshot", "File age")
		fmt.Printf("%s\n%s\n", header, strings.Repeat("-", len(header)))
		for idx, v := range scanResult.FileVersions {
			age := humanDuration(time.Since(v.Backup.MTime))
			fmt.Printf("%3d | %-[2]*s | %s\n", idx, width, v.Snapshot.Name, age)
		}

	case "diff":
		if len(flag.Args()) != 3 {
			fmt.Fprintf(os.Stderr, "Argument <#|SNAPSHOT> missing - use '%s -h'\n", os.Args[0])
			return
		}

		dr := scanner.NewDateRange(time.Now(), cliCfg.scanDays)
		sc := scanner.NewScanner(dr, "auto", ds, zfs)
		scanResult, err := sc.FindFileVersions(filePath)
		if err != nil {
			log.Errorf("scan failed - %v", err)
			return
		}


		var backup fs.FileHandle
		requestedVersion := flag.Arg(2)
		if idx, err := strconv.Atoi(requestedVersion); err == nil {
			// FIXME: bounds check
			backup = scanResult.FileVersions[idx].Backup
		} else {
			for _, v := range scanResult.FileVersions {
				if v.Snapshot.Name == requestedVersion {
					backup = v.Backup
					break
				}
			}
			log.Error("requested version not found")
			return
		}


		diffs, err := diff.NewDiffFromPath(backup.Path, filePath, 5)
		if err != nil {
			log.Errorf("unable to create diff - %v", err)
			return
		}

		fmt.Printf("%s", diffs.PrettyTextDiff)

	default:
		fmt.Fprintf(os.Stderr, "invalid action: %s\n", action)
		return
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

func parseFlags() (CliConfig, config.ZFSConfig) {
	cliCfg := new(CliConfig)

	// cli
	flag.BoolVar(&cliCfg.printVersion, "V", false, "print version and exit")
	flag.IntVar(&cliCfg.scanDays, "d", 7, "days to scan")

	// logging
	cliCfg.logLevel = plog.Note
	plog.FlagDebugVar(&cliCfg.logLevel, "v", "debug output")
	plog.FlagTraceVar(&cliCfg.logLevel, "vv", "trace output with caller location")

	// zfs
	zfsCfg := config.NewDefaultZFSConfig()
	flag.BoolVar(&zfsCfg.UseSudo, "use-sudo", zfsCfg.UseSudo, "use sudo when executing 'zfs' commands")
	flag.BoolVar(&zfsCfg.MountSnapshots, "mount-snapshots", zfsCfg.MountSnapshots,
		"mount snapshot (only necessary if it's not mounted by zfs automatically")

	flag.Parse()
	return *cliCfg, zfsCfg
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
