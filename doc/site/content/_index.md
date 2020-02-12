+++
title = "zfs-snap-diff"
type = "docs"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 10
+++

## Intro {#intro}

{{< hint danger >}}
This describes the currently ****unreleased alpha version****.
{{< /hint >}}

`zfs-snap-diff` searches different file versions in your zfs snapshots for you.

If you have hundreds or thousands of zfs snapshots, `zfs-snap-diff` searched
the snapshots and shows you only the snapshots where a given file was modified.

To speedup this process, it performs the search incremental when you request a older file version.

You can inspect a diff from the actual file version to the older file version in the
snapshot, revert a single change or restore a whole file.


## Usage {#usage}

-   install `zfs-snap-diff`

see: [Installation](/docs/install)

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

-   inspect a diff and revert to a older version

{{< figure src="/images/zfs-snap-diff.gif" alt="Example session from zfs-snap-diff" link="/images/zfs-snap-diff.gif" >}}
