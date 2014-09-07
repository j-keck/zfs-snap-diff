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
	zfsName       string
	zfsMountPoint string
)

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
	flag.Parse()

	if *printVersionFlag {
		fmt.Printf("Version: %s\n", VERSION)
		os.Exit(0)
	}

	// last argument is the zfs name
	zfsName = flag.Arg(0)

	// abort if zfs name is missing
	if len(zfsName) == 0 {
		fmt.Println("parameter <ZFS_NAME> missing\n")
		flag.Usage()
		os.Exit(1)
	}

	// lookup zfs mount point
	out, err := zfs("get -H -o value mountpoint " + zfsName)
	if err != nil {
		log.Print(out)
		os.Exit(1)
	}
	zfsMountPoint = out
	log.Printf("work on zfs: %s wich is mounted under: %s\n", zfsName, zfsMountPoint)

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

	// startup web server
	log.Printf("start server and listen on: '%s'\n", addr)
	listenAndServe(addr)
}
