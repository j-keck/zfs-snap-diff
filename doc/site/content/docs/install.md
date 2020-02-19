+++
title = "Installation"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 20
+++

{{< hint danger >}}
This describes the currently ****unreleased beta version****.
{{< /hint >}}


## Binary packages {#binary-packages}

You can download the latest binary package from here or
from the [githup release page](https://github.com/j-keck/zfs-snap-diff/releases).

{{<tabs "install">}}{{< tab "Linux" >}}
Download the beta version for ****Linux amd64**** here:

[zfs-snap-diff-linux-v1.0.0-beta-9-gc1980ab.tgz](/zfs-snap-diff-linux-v1.0.0-beta-9-gc1980ab.tgz)
{{< /tab >}}

{{< tab "FreeBSD" >}}
Download the beta version for ****FreeBSD amd64**** here:

[zfs-snap-diff-freebsd-v1.0.0-beta-9-gc1980ab.tgz](/zfs-snap-diff-freebsd-v1.0.0-beta-9-gc1980ab.tgz)
{{< /tab >}}

{{< tab "macOS" >}}
Download the beta version for ****macOS amd64**** here:

[zfs-snap-diff-darwin-v1.0.0-beta-9-gc1980ab.tgz](/zfs-snap-diff-darwin-v1.0.0-beta-9-gc1980ab.tgz)
{{< /tab >}}

{{< tab "Solaris" >}}
Download the beta version for ****Solaris amd64**** here:

[zfs-snap-diff-solaris-v1.0.0-beta-9-gc1980ab.tgz](/zfs-snap-diff-solaris-v1.0.0-beta-9-gc1980ab.tgz)
{{< /tab >}}

{{< /tabs >}}

{{< hint warning >}}
Try with the `-use-sudo` flag if it's not working - and please give feedback.
{{< /hint >}}

{{< hint info >}}
If you need a 32bit version, or a binary for a different
platform, feel free to contact me!
{{< /hint >}}


## Build from source {#build-from-source}

You need only [go](https://go-lang.org) to build this project.


### `go` {#go}

The minimum supported go version is `go1.12`.

-   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `go build -ldflags="-X main.version=$(git describe)" ./cmd/zfs-snap-diff`

The optional `-ldflags="-X main.version=$(git describe)"` flag updates the `version` string in the binary.


### `nix` {#nix}

I use [nix](https://nixos.org/nix/) to build my projects. The `nix` build also compiles the frontend
to javascript and decode it in `pkg/webapp/bindata.go`.

-   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `nix-build -A zfs-snap-diff`

The build artifacts `zfs-snap-diff` and `zsd` are in `./result/bin/`.

To crosscompile the binary use:

-   FreeBSD: `nix-build -A zfs-snap-diff --argstr goos freebsd`
-   MacOS: `nix-build -A zfs-snap-diff --argstr goos darwin`
-   Solaris: `nix-build -A zfs-snap-diff --argstr goos solaris`
