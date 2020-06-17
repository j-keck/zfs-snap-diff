+++
title = "Installation"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 20
+++

## Binary packages {#binary-packages}

You can download the latest binary package from ****here**** or from the [GitHub release page](https://github.com/j-keck/zfs-snap-diff/releases).

{{<tabs "install">}}
{{< tab "Linux" >}}
  1.) ****Download**** the latest version for ****Linux amd64****: [zfs-snap-diff-linux-v1.1.1.tgz](https://github.com/j-keck/zfs-snap-diff/releases/download/v1.1.1/zfs-snap-diff-linux-v1.1.1.tgz)

2.) Unpack the archive: `tar xvf zfs-snap-diff-linux-v1.1.1.tgz`

  3.) Run it:  `./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>`
{{< /tab >}}

{{< tab "FreeBSD" >}}
  1.) ****Download**** the latest version for ****FreeBSD amd64****: [zfs-snap-diff-freebsd-v1.1.1.tgz](https://github.com/j-keck/zfs-snap-diff/releases/download/v1.1.1/zfs-snap-diff-freebsd-v1.1.1.tgz)

2.) Unpack the archive: `tar xvf zfs-snap-diff-freebsd-v1.1.1.tgz`

  3.) Run it:  `./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>`
{{< /tab >}}

{{< tab "FreeBSD (pkg)" >}}You can use `pkg install zfs-snap-diff` to install it from the package repository.<br/>{{< hint info >}}The new 1.x.x series is currently only in the latest package set.{{< /hint >}}{{< /tab >}}
{{< tab "macOS" >}}
  1.) ****Download**** the latest version for ****macOS amd64****: [zfs-snap-diff-darwin-v1.1.1.tgz](https://github.com/j-keck/zfs-snap-diff/releases/download/v1.1.1/zfs-snap-diff-darwin-v1.1.1.tgz)

2.) Unpack the archive: `tar xvf zfs-snap-diff-darwin-v1.1.1.tgz`

  3.) Run it:  `./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>`
{{< /tab >}}

{{< tab "Solaris" >}}
  1.) ****Download**** the latest version for ****Solaris amd64****: [zfs-snap-diff-solaris-v1.1.1.tgz](https://github.com/j-keck/zfs-snap-diff/releases/download/v1.1.1/zfs-snap-diff-solaris-v1.1.1.tgz)

2.) Unpack the archive: `tar xvf zfs-snap-diff-solaris-v1.1.1.tgz`

  3.) Run it:  `./zfs-snap-diff [OPTIONS] <ZFS_DATASET_NAME>`
{{< /tab >}}

{{< /tabs >}}

{{< hint warning >}}
If you use any snapshot management functions, remember to use the `-use-sudo` flag!
{{< /hint >}}

{{< hint info >}}
Currently, the tar archive contains only the executables.
If you need distribution specific packages, or binaries for any other platform, feel free to [contact me](/docs/contact-support#contact).
{{< /hint >}}


## Build from source {#build-from-source}

The backend of `zfs-snap-diff` is implemented in [Go](https://golang.org), the frontend in [PureScript](http://purescript.org).


### `go` {#go}

I use [go-bindata](https://github.com/go-bindata/go-bindata) to decode the frontend code and all dependencies to a
go source file so you only need the go compiler to compile it yourself.

The minimum supported go version is `go1.12`.

-   clone this repo: `git clone --depth 1 https://github.com/j-keck/zfs-snap-diff`
-   `cd zfs-snap-diff`
-   build it: `go build -ldflags="-X main.version=$(git describe)" ./cmd/zfs-snap-diff`

The optional `-ldflags="-X main.version=$(git describe)"` flag updates the `version` string in the binary.


### `nix` {#nix}

I use [nix](https://nixos.org/nix/) to build my projects. The `nix` build also compiles the frontend
to javascript and decodes it in `pkg/webapp/bindata.go`.

-   clone this repo: `git clone --depth 1 https://github.com/j-keck/zfs-snap-diff`
-   change to the checkout directory: `cd zfs-snap-diff`
-   build it: `nix-build -A zfs-snap-diff`

The build artifacts `zfs-snap-diff` and `zsd` are in `./result/bin/`.

To crosscompile the binary use:

-   FreeBSD: `nix-build -A zfs-snap-diff --argstr goos freebsd`
-   MacOS: `nix-build -A zfs-snap-diff --argstr goos darwin`
-   Solaris: `nix-build -A zfs-snap-diff --argstr goos solaris`
