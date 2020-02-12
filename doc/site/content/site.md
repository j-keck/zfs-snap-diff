+++
title = "zfs-snap-diff"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
+++

## Index {#index}


### Intro {#intro}

{{< hint danger >}}
This describes the currently ****unreleased alpha version**** of \`zfs-snap-diff\`.
{{< /hint >}}

`zfs-snap-diff` searches different file versions in your zfs snapshots for you.

If you have hundreds or thousands of zfs snapshots, `zfs-snap-diff` searched
the snapshots and shows you only the snapshots where a given file was modified.

To speedup this process, it performs the search incremental when you request a older file version.

You can inspect a diff from the actual file version to the older file version in the
snapshot, revert a single change or restore a whole file.


### Usage {#usage}

-   install `zfs-snap-diff` see: [Installation](/docs/install)

-   startup the daemon

<!--listend-->

```sh
./zfs-snap-diff <ZFS_DATASET_NAME>
```

-   open your webbrowser at

<!--listend-->

```sh
http://127.0.0.1:12345
```

{{< figure src="/images/browse-filesystem.png" alt="Screenshot from 'Browse filesystem'" link="/images/browse-filesystem.png" >}}


## Installation {#installation}


#### Linux {#linux}

TODO


#### FreeBSD {#freebsd}

TODO


#### Build from source {#build-from-source}

The minimum supported go version is `go1.12`.

-    `go`

    -   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
    -   change to the checkout directory: `cd zfs-snap-diff`
    -   build it: `go build -ldflags="-X main.version=$(git describe)" ./cmd/zfs-snap-diff`

    The optional `-ldflags="-X main.version=$(git describe)"` flag updates the `version` string in the binary.

-    `nix`

    The `nix` build also compiles the frontend to javascript and decode it in `pkg/webapp/bindata.go`.

    -   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
    -   change to the checkout directory: `cd zfs-snap-diff`
    -   build it: `nix-build -A zfs-snap-diff`

    To crosscompile the binary for

    -   FreeBSD: `nix-build -A zfs-snap-diff --argstr goos freebsd`
    -   MacOS: `nix-build -A zfs-snap-diff --argstr goos darwin`
    -   Solaris: `nix-build -A zfs-snap-diff --argstr goos solaris`


## User Guide {#user-guide}


### Browse the actual filesytem {#browse-the-actual-filesytem}

{{< figure src="/images/browse-filesystem.png" alt="Screenshot from 'Browse filesystem'" link="/images/browse-filesystem.png" >}}


### Browse snapshots {#browse-snapshots}

{{< figure src="/images/browse-snapshots.png" alt="Screenshot from 'Browse snapshots" link="/images/browse-snapshots.png" >}}


## Changelog {#changelog}


### 1.0.0 (unreleased) {#1-dot-0-dot-0--unreleased}

This version is a complete rewrite

-   date-range based search for file versions
    -   this speeds up the scan dramatically if
        there are thousands snapshots on spinning disk

-   bookmarks

-   works now also with 'legacy' mountpoints

-   new backend and frontend code
