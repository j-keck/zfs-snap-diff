+++
title = "Installation"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 20
+++

{{< hint danger >}}
This describes the currently ****unreleased beta version****.
{{< /hint >}}

If you need a 32bit version, or a binary for a different
platform, feel free to contact me!

The tgz archive contains

-   [zfs-snap-diff](/docs/guide/#zfs-snap-diff): web based tool to find different file versions in your zfs-snapshots.

-   [zsd](/docs/guide/#zsd): a little independent cli tool to find differnt file versions in your zfs-snapshots.


## Linux {#linux}

Download the alpha version for Linux amd64 here:

[zfs-snap-diff-linux-v1.0.0-beta.tgz](/zfs-snap-diff-linux-v1.0.0-beta.tgz)

<span class="underline">Try with the \`-use-sudo\` if it's not working - and please give feedbak if somethink is not working</span>


## FreeBSD {#freebsd}

Download the alpha version for FreeBSD amd64 here:

[zfs-snap-diff-freebsd-v1.0.0-beta.tgz](/zfs-snap-diff-freebsd-v1.0.0-beta.tgz)

<span class="underline">Try with the \`-use-sudo\` if it's not working - and please give feedbak if somethink is not working</span>


## macOS {#macos}

Download the alpha version for macOS amd64 here:

[zfs-snap-diff-darwin-v1.0.0-beta.tgz](/zfs-snap-diff-darwin-v1.0.0-beta.tgz)

<span class="underline">Try with the \`-use-sudo\` if it's not working - and please give feedbak if somethink is not working</span>


## Solaris {#solaris}

Download the alpha version for Solaris amd64 here:

[zfs-snap-diff-solaris-v1.0.0-beta.tgz](/zfs-snap-diff-solaris-v1.0.0-beta.tgz)

<span class="underline">Try with the \`-use-sudo\` if it's not working - and please give feedbak if somethink is not working</span>


## Build from source {#build-from-source}


### `go` {#go}

The minimum supported go version is `go1.12`.

-   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `go build -ldflags="-X main.version=$(git describe)" ./cmd/zfs-snap-diff`

The optional `-ldflags="-X main.version=$(git describe)"` flag updates the `version` string in the binary.


### `nix` {#nix}

I use [nix](https://nixos.org/nix/) to build my projects.

The `nix` build also compiles the frontend to javascript and decode it in `pkg/webapp/bindata.go`.

-   clone this repo: `git clone -b dev https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `nix-build -A zfs-snap-diff`

The build artifacts `zfs-snap-diff` and `zsd` are in `./result/bin/`.

To crosscompile the binary use:

-   FreeBSD: `nix-build -A zfs-snap-diff --argstr goos freebsd`
-   MacOS: `nix-build -A zfs-snap-diff --argstr goos darwin`
-   Solaris: `nix-build -A zfs-snap-diff --argstr goos solaris`
