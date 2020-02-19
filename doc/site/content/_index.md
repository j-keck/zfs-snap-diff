+++
title = "zfs-snap-diff"
type = "docs"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 10
+++

{{< hint danger >}}
This describes the currently ****unreleased beta version****.
{{< /hint >}}

`zfs-snap-diff` searches different file versions in your zfs snapshots for you.

If you have hundreds or thousands of zfs snapshots, `zfs-snap-diff` searched
the snapshots and shows you only the snapshots where a given file was modified.

To speedup this process, it performs the search incremental when you request an older file version.

You can inspect a diff from the actual file version to the older file version in the
snapshot, revert a single change or restore a whole file.

`zfs-snap-diff` has a web frontend, so it can run on your local work machine or on your
remote file / backup server (no Xserver necesarry). To keep it portable it's made
as a single static compiled executable.

{{< figure src="/images/zfs-snap-diff.gif" alt="Example session from zfs-snap-diff" link="/images/zfs-snap-diff.gif" >}}
