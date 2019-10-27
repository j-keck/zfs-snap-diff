package main

import (
	"flag"
	"fmt"
	"github.com/j-keck/plog"
	"github.com/j-keck/zfs-snap-diff/pkg/oldweb"
	"github.com/j-keck/zfs-snap-diff/pkg/zfs"
	"os"
)

// VERSION  a set at buid time (go build -ldflags "-X main.VERSION $(git describe)")
var VERSION string

func main() {
	log := plog.GlobalLogger().Add(plog.NewDefaultConsoleLogger())

	// formate help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, " Usage\n=======\n%s [OPT_ARGS] <ZFS_NAME>\n", os.Args[0])
		fmt.Fprint(os.Stderr, "OPT_ARGS:\n")
		flag.PrintDefaults()
	}

	// define flags / parse flags
	addrFlag := flag.String("l", "127.0.0.1", "web server listen address")
	portFlag := flag.Int("p", 12345, "web server port")
	useTLSFlag := flag.Bool("tls", false, "use TLS - NOTE: -cert <CERT_FILE> -key <KEY_FILE> are mandatory")
	certFileFlag := flag.String("cert", "", "certificate file for TLS")
	keyFileFlag := flag.String("key", "", "private key file for TLS")
	listenOnAllInterfacesFlag := flag.Bool("a", false, "listen on all interfaces")
	printVersionFlag := flag.Bool("V", false, "print version and exit")
	useSudoFlag := flag.Bool("use-sudo", false, "use sudo when executing os commands")

	// default log level
	logLevel := plog.Info
	plog.FlagDebugVar(&logLevel, "v", "verbose logging")
	plog.FlagTraceVar(&logLevel, "vv", "trace logging")

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

	log.SetLevel(logLevel)
	log.Debugf("zfs-snap-diff version: %s", VERSION)

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
	if *useTLSFlag {
		if len(*certFileFlag) == 0 || len(*keyFileFlag) == 0 {
			fmt.Println("ABORT: parameter -cert <CERT_FILE> -key <KEY_FILE> are mandatory")
			os.Exit(1)
		}

		if _, err := os.Stat(*certFileFlag); os.IsNotExist(err) {
			fmt.Printf("ABORT: cert file '%s' not found\n", *certFileFlag)
			os.Exit(1)
		}

		if _, err := os.Stat(*keyFileFlag); os.IsNotExist(err) {
			fmt.Printf("ABORT: key file '%s' not found\n", *keyFileFlag)
			os.Exit(1)
		}
	}

	zfs, err := zfs.NewZFS(zfsName, *useSudoFlag)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// listen on the given address - or if flag '-a' is given, listen on all interfaces
	var addr string
	if *listenOnAllInterfacesFlag {
		fmt.Println("")
		fmt.Println("!! ** WARNING **                !!")
		fmt.Println("!! LISTEN ON ALL INTERFACES     !!")
		fmt.Println("!! CURRENTLY NO AUTHENTICATION  !!")
		if !*useTLSFlag {
			fmt.Println("\nHINT: USE -tls -cert <CERT_FILE> -key <KEY_FILE> to enable encryption!")
		}
		fmt.Println("")
		addr = fmt.Sprintf("0.0.0.0:%d", *portFlag)
	} else {
		addr = fmt.Sprintf("%s:%d", *addrFlag, *portFlag)
	}

	// print warning if file-compare method md5 is used
	if *compareFileMethodFlag == "md5" {
		if *scanSnapLimitFlag > 0 {
			log.Warn("compare all files only with md5 - expect high cpu usage / long runtime!")
		} else {
			log.Warn("no 'scan-snap-limit' was given and compare all files only with md5 - expect VERY HIGH cpu usage / VERY LONG runtime!!!!")
		}
	}

	// webserver config
	webServerCfg := oldweb.WebServerConfig{
		addr,
		*useTLSFlag,
		*certFileFlag,
		*keyFileFlag,
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
	oldweb.ListenAndServe(zfs, webServerCfg, frontendCfg)
}
