+++
title = "User Guide"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 30
+++

{{< hint danger >}}
This describes the currently ****unreleased beta version****.
{{< /hint >}}


## `zfs-snap-diff` {#zfs-snap-diff}

With `zfs-snap-diff` you can

-   find older file versions in your zfs-snapshots for a given file

-   inspect a diff, revert a single change, restore the whole file
    or download the older version

-   browse the directory content from a snapshot

-   download a zip-archive from any folder in your snapshots

-   create and destroy snapshots in the webapp

-   bookmark often used folders

<!--listend-->

```text
USAGE:
  ./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>

OPTIONS:
  -V	print version and exit
  -a	listen on all interfaces
  -cert string
        TLS certificate file
  -d int
        days to scan (default 7)
  -key string
        TLS private key file
  -l string
        webserver listen address (default "127.0.0.1")
  -log-locations
        log messages with caller location
  -log-timestamps
        log messages with timestamps in unix format
  -mount-snapshots
        mount snapshot (only necessary if it's not mounted by zfs automatically
  -p int
        webserver port (default 12345)
  -tls
        use TLS - NOTE: -cert <CERT_FILE> -key <KEY_FILE> are mandatory
  -use-cache-dir-for-backups
        use platform depend user local cache directory for backups (default true)
  -use-sudo
        use sudo when executing 'zfs' commands
  -v	debug output
  -vv
        trace output with caller location
  -webapp-dir string
        when given, serve the webapp from the given directory
```


### Browse the actual filesytem {#browse-the-actual-filesytem}

You can browse the actual filesystem and inspect a diff from the actual file version to the older
file version in the selected snapshot, revert a single change or restore a whole file.

{{< figure src="/images/browse-filesystem.png" alt="Screenshot from 'Browse filesystem'" link="/images/browse-filesystem.png" >}}


### Browse snapshots {#browse-snapshots}

In this view you can view all snapshots.

{{< figure src="/images/browse-snapshots-snapshots.png" alt="Screenshot from 'Browse snapshots'" link="/images/browse-snapshots-snapshots.png" >}}

and inspect the directory content where the snapshot was created

{{< figure src="/images/browse-snapshots-dir-browser.png" alt="Browse snapshots / directory browser" link="/images/browse-snapshots-dir-browser" >}}


## `zsd` {#zsd}

`zsd` uses the same code to find different file versions in your snapshots,
diff / view a older file version or restore a file from a given snapshot.

```text
zsd is a little independent cli tool to find different file versions in your zfs-snapshots.

USAGE:
 ./zsd [OPTIONS] <FILE> <ACTION>

OPTIONS:
  -V	print version and exit
  -d int
        days to scan (default 7)
  -v	debug output
  -vv
        trace output with caller location

ACTIONS:
  list                : list zfs-snapshots with different file-versions for the given file
  cat     <#|SNAPSHOT>: show the file-content from the given snapshot
  diff    <#|SNAPSHOT>: show differences between the actual version and from the selected snapshot
  restore <#|SNAPSHOT>: restore the file from the given snapshot
```

-   list zfs-snapshots where the given file was modified

<!--listend-->

```text
main⟩ ./zsd go.mod list
scan the last 7 days for other file versions
  # | Snapshot                               | Snapshot age
-----------------------------------------------------------
  0 | zfs-auto-snap_hourly-2020-02-12-12h00U | 5 hours
  1 | zfs-auto-snap_hourly-2020-02-12-09h00U | 8 hours
```

-   show the file-content from the given snapshot

<!--listend-->

```text
 main⟩ ./zsd go.mod cat 0
module github.com/j-keck/zfs-snap-diff

require (
	github.com/j-keck/go-diff v1.0.0
	github.com/j-keck/plog v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
)

go 1.12
```

-   show the differences between the actual version and from the given snapshot

<!--listend-->

```text
main⟩ ./zsd go.mod diff 0
Diff from the actual version to the version from: 2020-02-12 10:07:44.434355182 +0100 CET
module github.com/j-keck/zfs-snap-diff

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/j-keck/go-diff v1.0.0
-	github.com/j-keck/plog v0.5.0
+	github.com/j-keck/plog v0.6.0
	github.com/stretchr/testify v1.4.0 // indirect
)

go 1.12
```

-   restore the given file with an older version

<!--listend-->

```text
main⟩ ./zsd go.mod restore 0
backup from the actual version created at: /home/j/.cache/zfs-snap-diff/backups/home/j/prj/priv/zfs-snap-diff/go.mod_20200212_182709%
version restored from snapshot: zfs-auto-snap_hourly-2020-02-12-12h00U
```
