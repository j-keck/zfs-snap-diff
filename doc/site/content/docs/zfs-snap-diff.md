+++
title = "zfs-snap-diff"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 30
+++

{{< hint danger >}}
This describes the currently ****unreleased beta version****.
{{< /hint >}}

`zfs-snap-diff` - web application to find older versions of a given file in your zfs snapshots.

With `zfs-snap-diff` you can

-   find older file versions in your zfs-snapshots for a given file

-   view the file content from a given snapshot

-   inspect a diff from the older version to the actual version

-   revert a single change

-   restore a version from a zfs snapshot

-   download a file version

-   browse the directory content from a snapshot

-   download a zip-archive from any folder in your snapshots

-   create and destroy snapshots in the webapp

-   bookmark often used folders


## Usage {#usage}

```text
main⟩ zfs-snap-diff -h
zfs-snap-diff - web application to find older versions of a given file in your zfs snapshots.

USAGE:
  ./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>

OPTIONS:
  -V	print version and exit
  -a	listen on all interfaces
  -cert string
        TLS certificate file
  -compare-method string
        used method to determine if a file was modified ('auto', 'mtime', 'size+mtime', 'content', 'md5') (default "auto")
  -d int
        days to scan (default 7)
  -diff-context-size int
        show N lines before and after each diff (default 5)
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

Project home page: https://j-keck.github.io/zfs-snap-diff
```


## Startup {#startup}

-   startup a server instance

<!--listend-->

```sh
./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>
```

This starts a embedded webserver and serves the included web-app.

-   open your webbrowser at

<!--listend-->

```sh
http://127.0.0.1:12345
```


## Browse the actual filesytem {#browse-the-actual-filesytem}

You can browse the actual filesystem and inspect a diff from the actual file version to the older
file version in the selected snapshot, revert a single change or restore a whole file.

{{< figure src="/images/browse-filesystem.png" alt="Screenshot from 'Browse filesystem'" link="/images/browse-filesystem.png" >}}


## Browse snapshots {#browse-snapshots}

In this view you can view all snapshots.

{{< figure src="/images/browse-snapshots-snapshots.png" alt="Screenshot from 'Browse snapshots'" link="/images/browse-snapshots-snapshots.png" >}}

and inspect the directory content where the snapshot was created

{{< figure src="/images/browse-snapshots-dir-browser.png" alt="Browse snapshots / directory browser" link="/images/browse-snapshots-dir-browser" >}}


## Create snapshots {#create-snapshots}

The default snapshot template name is configurable: `snapshot-name-template`
  [![](/images/create-snapshot.png)](/images/create-snapshot.png)


## Configuration {#configuration}

TODO
