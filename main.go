package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// VERSION  a set at buid time (go build -ldflags "-X main.VERSION $(git describe)")
var VERSION string

// ZFS Handler
var zfs *ZFS

// Log Handler
var (
	logDebug  *log.Logger
	logInfo   *log.Logger
	logNotice *log.Logger
	logWarn   *log.Logger
	logError  *log.Logger
)

// FrontendConfig hold the configuration for the ui
type FrontendConfig map[string]interface{}

func main() {
	// formate help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, " Usage\n=======\n%s [OPT_ARGS] <ZFS_NAME>\n", os.Args[0])
		fmt.Fprint(os.Stderr, "OPT_ARGS:\n")
		flag.PrintDefaults()
	}

	// define flags / parse flags
	portFlag := flag.Int("p", 12345, "web server port")
	listenOnAllInterfacesFlag := flag.Bool("a", false, "listen on all interfaces")
	printVersionFlag := flag.Bool("V", false, "print version and exit")
	verboseLoggingFlag := flag.Bool("v", false, "verbose logging")
	useSudoFlag := flag.Bool("use-sudo", false, "use sudo when executing os commands")

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

	// init logging handler
	if *verboseLoggingFlag {
		initLogHandlers(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		initLogHandlers(ioutil.Discard, os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	}

	// last argument is the zfs name
	zfsName := flag.Arg(0)

	// abort if zfs name is missing
	if len(zfsName) == 0 {
		fmt.Println("parameter <ZFS_NAME> missing")
		fmt.Println()
		flag.Usage()
		os.Exit(1)
	}

	// initialize zfs handler
	var err error
	zfs, err = NewZFS(zfsName, *useSudoFlag)
	if err != nil {
		logError.Println(err.Error())
		os.Exit(1)
	}
	logInfo.Printf("work on zfs: %s which is mounted under: %s\n", zfs.Datasets.Root().Name, zfs.Datasets.Root().MountPoint)

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
			logNotice.Println("compare all files only with md5 - expect high cpu usage / long runtime!")
		} else {
			logNotice.Println("no 'scan-snap-limit' was given and compare all files only with md5 - expect VERY HIGH cpu usage / VERY LONG runtime!!!!")
		}
	}

	// frontend-config
	frontendConfig := FrontendConfig{
		"diffContextSize":   *diffContextSizeFlag,
		"defaultFileAction": *defaultFileActionFlag,
		"compareFileMethod": *compareFileMethodFlag,
		"datasets":          zfs.Datasets,
	}
	if *scanSnapLimitFlag >= 0 {
		// only add positive values - negative values: scan all snapshots
		frontendConfig["scanSnapLimit"] = *scanSnapLimitFlag
	}

	// startup web server
	logInfo.Printf("start server and listen on: '%s'\n", addr)
	listenAndServe(addr, frontendConfig)
}

func initLogHandlers(debugHndl, infoHndl, noticeHndl, warnHndl, errorHndl io.Writer) {
	logDebug = log.New(debugHndl, "DEBUG:  ", log.Ldate|log.Ltime|log.Lshortfile)
	logInfo = log.New(infoHndl, "INFO:   ", log.Ldate|log.Ltime)
	logNotice = log.New(noticeHndl, "NOTICE: ", log.Ldate|log.Ltime)
	logWarn = log.New(warnHndl, "WARN:   ", log.Ldate|log.Ltime)
	logError = log.New(errorHndl, "ERROR:  ", log.Ldate|log.Ltime|log.Lshortfile)
}
