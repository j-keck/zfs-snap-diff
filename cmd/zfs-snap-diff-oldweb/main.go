package main

import (
	"flag"
	"fmt"
	"github.com/j-keck/zfs-snap-diff/pkg/oldweb"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"github.com/j-keck/plog"
	"os"
)

// VERSION  a set at buid time (go build -ldflags "-X main.VERSION $(git describe)")
var VERSION string

func main() {
	// formate help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, " Usage\n=======\n%s [OPT_ARGS] <ZFS_NAME>\n", os.Args[0])
		fmt.Fprint(os.Stderr, "OPT_ARGS:\n")
		flag.PrintDefaults()
	}

	// frontend config
	diffContextSizeFlag := flag.Int("diff-context-size", 5, "context size in diff")
	defaultFileActionFlag := flag.String("default-file-action", "view",
		"default file action in frontend when a file is selected: 'off', 'view', 'diff', 'download', 'restore'")

	scanSnapLimitFlag := flag.Int("scan-snap-limit", -1,
		"scan snapshots where file was modified limit (negative values: scan all snapshots)")
	compareFileMethodFlag := flag.String("compare-file-method", "auto",
		"compare method when searching snapshots for other file versions: 'auto', 'size+modTime', 'size' or 'md5'")

	// parse config
	cliCfg, zsdCfg := parseFlags()
	setupLogger(cliCfg)
	log := plog.GlobalLogger()

	if cliCfg.printVersion {
		fmt.Printf("zfs-snap-diff: %s\n", VERSION)
		return
	}


	// last argument is the zfs name
	zfsName := flag.Arg(0)

	// abort if zfs name is missing
	if len(zfsName) == 0 {
		fmt.Println("ABORT: parameter <ZFS_NAME> missing")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	// validate args for tls
	if zsdCfg.Webserver.UseTLS {
		if len(zsdCfg.Webserver.CertFile) == 0 || len(zsdCfg.Webserver.KeyFile) == 0 {
			fmt.Println("ABORT: parameter -cert <CERT_FILE> -key <KEY_FILE> are mandatory")
			os.Exit(1)
		}

		if _, err := os.Stat(zsdCfg.Webserver.CertFile); os.IsNotExist(err) {
			fmt.Printf("ABORT: cert file '%s' not found\n", zsdCfg.Webserver.CertFile)
			os.Exit(1)
		}

		if _, err := os.Stat(zsdCfg.Webserver.KeyFile); os.IsNotExist(err) {
			fmt.Printf("ABORT: key file '%s' not found\n", zsdCfg.Webserver.KeyFile)
			os.Exit(1)
		}
	}

	zfs, err := zfs.NewZFS(zfsName, zsdCfg)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if zsdCfg.Webserver.ListenOnAllInterfaces {
		fmt.Println("")
		fmt.Println("!! ** WARNING **                !!")
		fmt.Println("!! LISTEN ON ALL INTERFACES     !!")
		fmt.Println("!! CURRENTLY NO AUTHENTICATION  !!")
		if !zsdCfg.Webserver.UseTLS {
			fmt.Println("\nHINT: USE -tls -cert <CERT_FILE> -key <KEY_FILE> to enable encryption!")
		}
		fmt.Println("")
	}

	// print warning if file-compare method md5 is used
	if *compareFileMethodFlag == "md5" {
		if *scanSnapLimitFlag > 0 {
			log.Warn("compare all files only with md5 - expect high cpu usage / long runtime!")
		} else {
			log.Warn("no 'scan-snap-limit' was given and compare all files only with md5 - expect VERY HIGH cpu usage / VERY LONG runtime!!!!")
		}
	}


	// frontend-config
	frontendCfg := oldweb.FrontendConfig{
		"diffContextSize":   *diffContextSizeFlag,
		"defaultFileAction": *defaultFileActionFlag,
		"compareFileMethod": *compareFileMethodFlag,
		"datasets":          zfs.Datasets(),
	}
	if *scanSnapLimitFlag >= 0 {
		// only add positive values - negative values: scan all snapshots
		frontendCfg["scanSnapLimit"] = *scanSnapLimitFlag
	}

	// startup web server
	oldweb.ListenAndServe(zfs, zsdCfg.Webserver, frontendCfg)
}
