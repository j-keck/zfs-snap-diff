package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// VERSION  a set at buid time (go build -ldflags "-X main.VERSION $(git describe)")
var VERSION string

var (
	zfs *ZFS
)

// FrontendConfig hold the configuration for the ui
type FrontendConfig map[string]interface{}

func main() {
	// formate help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, " Usage\n=======\n%s [OPT_ARGS] <ZFS_NAME>\n\n", os.Args[0])
		fmt.Fprint(os.Stderr, "OPT_ARGS:\n")
		flag.PrintDefaults()
	}

	// define flags / parse flags
	portFlag := flag.Int("p", 12345, "web server port")
	listenOnAllInterfacesFlag := flag.Bool("a", false, "listen on all interfaces")
	printVersionFlag := flag.Bool("V", false, "print version and exit")
	// frontend
	diffContextSizeFlag := flag.Int("diff-context-size", 5, "context size in diff")
	defaultFileActionFlag := flag.String("default-file-action", "view", "default file action in frontend when a file is selected: 'off', 'view', 'diff', 'download', 'restore'")

	scanSnapLimitFlag := flag.Int("scan-snap-limit", -1, "scan snapshots where file was modified limit (negative values: scan all snapshots)")
	compareFileMethodFlag := flag.String("compare-file-method", "auto", "compare method when searching snapshots for other file versions: 'auto', 'size+modTime', 'size' or 'md5'")

	flag.Parse()

	if *printVersionFlag {
		fmt.Printf("Version: %s\n", VERSION)
		os.Exit(0)
	}

	// last argument is the zfs name
	zfsName := flag.Arg(0)

	// abort if zfs name is missing
	if len(zfsName) == 0 {
		fmt.Println("parameter <ZFS_NAME> missing\n")
		flag.Usage()
		os.Exit(1)
	}

	// initialize zfs handler
	var err error
	zfs, err = NewZFS(zfsName)
	if err != nil {
		log.Print(err.Error())
		os.Exit(1)
	}
	log.Printf("work on zfs: %s wich is mounted under: %s\n", zfs.Name, zfs.MountPoint)

	// listen on localhost - if flag '-a' is given, listen on all interfaces
	var addr string
	if *listenOnAllInterfacesFlag {
		fmt.Println("")
		fmt.Println("!! ** WARNING **                            !!")
		fmt.Println("!! LISTEN ON ALL INTERFACES                 !!")
		fmt.Println("!! CURRENTLY NO ENCRYPTION / AUTHENTICATION !!")
		fmt.Println("")
		addr = fmt.Sprintf(":%d", *portFlag)
	} else {
		addr = fmt.Sprintf("127.0.0.1:%d", *portFlag)
	}

	// print warning if file-compare method md5 is used
	if *compareFileMethodFlag == "md5" {
		if *scanSnapLimitFlag > 0 {
			log.Println("NOTICE: compare all files only with md5 - expect high cpu usage!")
		} else {
			log.Println("WARNING: no 'scan-snap-limit' was given and compare file with md5 - expect VERY HIGH cpu usage / VERY LONG runtime!!!!")
		}
	}

	// frontend-config
	frontendConfig := FrontendConfig{
		"zfsMountPoint":     zfs.MountPoint,
		"diffContextSize":   *diffContextSizeFlag,
		"defaultFileAction": *defaultFileActionFlag,
		"compareFileMethod": *compareFileMethodFlag,
	}
	if *scanSnapLimitFlag >= 0 {
		// only add positive values - negative values: scan all snapshots
		frontendConfig["scanSnapLimit"] = *scanSnapLimitFlag
	}

	// startup web server
	log.Printf("start server and listen on: '%s'\n", addr)
	listenAndServe(addr, frontendConfig)
}
