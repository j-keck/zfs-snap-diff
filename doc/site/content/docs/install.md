+++
title = "Installation"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 20
+++

{{< hint danger >}}
This describes the currently ****unreleased alpha version****.
{{< /hint >}}

If you need a 32bit version, or a binary for a different
platform, feel free to contact me!


## Linux {#linux}

Download the alpha version for Linux amd64 from:

[zfs-snap-diff-linux-v1.0.0-alpha-9-gcdce9fb.gz](/zfs-snap-diff-linux-v1.0.0-alpha-9-gcdce9fb.gz)

If the `zfs` executable is not in your $PATH, try `-use-sudo`.


## FreeBSD {#freebsd}

Download the alpha version for FreeBSD amd64 from:

[zfs-snap-diff-freebsd-v1.0.0-alpha-9-gcdce9fb.gz](/zfs-snap-diff-freebsd-v1.0.0-alpha-9-gcdce9fb.gz)


## macOS {#macos}

Download the alpha version from:

[zfs-snap-diff-macos-v1.0.0-alpha-9-gcdce9fb.gz](/zfs-snap-diff-macos-v1.0.0-alpha-9-gcdce9fb.gz)

_Try whith `-use-sudo` if it's not working - and please give feedbak if it works (or not)_


## Build from source {#build-from-source}


### `go` {#go}

The minimum supported go version is `go1.12`.

-   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `go build -ldflags="-X main.version=$(git describe)" ./cmd/zfs-snap-diff`

The optional `-ldflags="-X main.version=$(git describe)"` flag updates the `version` string in the binary.


### `nix` {#nix}

The `nix` build also compiles the frontend to javascript and decode it in `pkg/webapp/bindata.go`.

-   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `nix-build -A zfs-snap-diff`

The build artifacts `zfs-snap-diff` and `zsd` are in `./result/bin/`.

To crosscompile the binary use:

-   FreeBSD: `nix-build -A zfs-snap-diff --argstr goos freebsd`
-   MacOS: `nix-build -A zfs-snap-diff --argstr goos darwin`
-   Solaris: `nix-build -A zfs-snap-diff --argstr goos solaris`
