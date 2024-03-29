+++
title = "zfs-snap-diff"
type = "docs"
draft = false
creator = "Emacs 27.2 (Org mode 9.4.4 + ox-hugo)"
weight = 10
+++

`zfs-snap-diff` helps you with your zfs snapshots.


## Find file versions {#find-file-versions}

`zfs-snap-diff` searches different file versions in your zfs snapshots for you.

If you have hundreds or thousands of zfs snapshots, `zfs-snap-diff` searches through
the snapshots and shows you only the snapshots where a given file was modified.

You can inspect a diff from the actual file version to the older file version in the
snapshot, revert a single change or restore a whole file.


## Management {#management}

You can create, destroy, rename, rollback and clone zfs snapshots, use the integrated directory browser to
navigate in your snapshots (directory history) and download a directory as a zip-archive.


## Simple use {#simple-use}

[`zfs-snap-diff`](/docs/zfs-snap-diff) has a web frontend, so it can run on your local work machine or on your
remote file / backup server (no Xserver necessary). To keep it portable it's made
as a single static compiled executable.

_For a terminal based tool see [zsd](https://j-keck.github.io/zsd)._

{{< figure src="/images/zfs-snap-diff.gif" alt="Example session from zfs-snap-diff" link="/images/zfs-snap-diff.gif" >}}


## License {#license}

`zfs-snap-diff` is released under the ****MIT**** License. See the [license file](https://github.com/j-keck/zfs-snap-diff/blob/master/LICENSE) for more information.
