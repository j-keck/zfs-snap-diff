+++
title = "zsd"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 35
+++

`zsd` - cli tool to find older versions of a given file in your zfs snapshots.

With `zsd` you can

-   find older file versions in your zfs snapshots for a given file

-   view the file content from a given snapshot

-   inspect a diff from the older version to the actual version

-   restore a version from a zfs snapshot

It uses the same code as `zfs-snap-diff` to find different file versions in your
zfs snapshots.


## Usage {#usage}

```text
main⟩ zsd -h
zsd - cli tool to find older versions of a given file in your zfs snapshots.

USAGE:
 ./zsd [OPTIONS] <FILE> <ACTION>

OPTIONS:
  -V	print version and exit
  -d int
        days to scan (default 2)
 -mount-snapshots
        mount snapshot (only necessary if it's not mounted by zfs automatically)
 -use-sudo
        use sudo when executing 'zfs' commands
  -v	debug output
  -vv
        trace output with caller location

ACTIONS:
  list                : list zfs snapshots where the given file was modified
  cat     <#|SNAPSHOT>: show file content of the given snapshot
  diff    <#|SNAPSHOT>: show a diff from the selected snapshot to the actual version
  restore <#|SNAPSHOT>: restore the file from the given snapshot

You can use the snapshot number from the `list` output or the snapshot name to select a snapshot.

Project home page: https://j-keck.github.io/zfs-snap-diff
```


## List snapshots {#list-snapshots}

Use the `list` action to list all snapshots where the
given file was modified.

```text
main⟩ zsd go.mod list
scan the last 7 days for other file versions
  # | Snapshot                               | Snapshot age
-----------------------------------------------------------
  0 | zfs-auto-snap_hourly-2020-02-12-12h00U | 5 hours
  1 | zfs-auto-snap_hourly-2020-02-12-09h00U | 8 hours
```


## Show file content {#show-file-content}

Use the `cat` action to show the file content from
the given snapshot.

{{< hint info >}}
You can use the snapshot number from the `list` output
or the snapshot name to select a snapshot.
{{< /hint >}}

```text
main⟩ zsd go.mod cat 0
module github.com/j-keck/zfs-snap-diff

require (
	github.com/j-keck/go-diff v1.0.0
	github.com/j-keck/plog v0.5.0
	github.com/stretchr/testify v1.4.0 // indirect
)

go 1.12
```


## Show diff {#show-diff}

To show a diff from the selected snapshot to the actual version
use the `diff` action.

{{< hint info >}}
You can use the snapshot number from the `list` output
or the snapshot name to select a snapshot.
{{< /hint >}}

```text
main⟩ zsd go.mod diff 0
Diff from the actual version to the version from: 2020-02-12 10:07:44.434355182 +0100 CET
module github.com/j-keck/zfs-snap-diff

require (
   github.com/BurntSushi/toml v0.3.1
   github.com/j-keck/go-diff v1.0.0
-  github.com/j-keck/plog v0.5.0
+  github.com/j-keck/plog v0.6.0
   github.com/stretchr/testify v1.4.0 // indirect
)

go 1.12
```


## Restore file {#restore-file}

To restore a given file with an older version use `restore`.

{{< hint info >}}
You can use the snapshot number from the `list` output
or the snapshot name to select a snapshot.
{{< /hint >}}

```text
main⟩ zsd go.mod restore 0
backup from the actual version created at: /home/j/.cache/zfs-snap-diff/backups/home/j/prj/priv/zfs-snap-diff/go.mod_20200212_182709%
version restored from snapshot: zfs-auto-snap_hourly-2020-02-12-12h00U
```

{{< hint warning >}}
A backup of the current version will be created.
{{< /hint >}}
