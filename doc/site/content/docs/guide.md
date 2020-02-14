+++
title = "User Guide"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 30
+++

{{< hint danger >}}
This describes the currently ****unreleased alpha version****.
{{< /hint >}}


## `zfs-snap-diff` {#zfs-snap-diff}


### Browse the actual filesytem {#browse-the-actual-filesytem}

You can browse the actual filesystem an inspect a diff from the actual file version to the older
file version in the selected snapshot, revert a single change or restore a whole file.

{{< figure src="/images/browse-filesystem.png" alt="Screenshot from 'Browse filesystem'" link="/images/browse-filesystem.png" >}}


### Browse snapshots {#browse-snapshots}

In this view you can view the content of your snapshots.

{{< figure src="/images/browse-snapshots.png" alt="Screenshot from 'Browse snapshots" link="/images/browse-snapshots.png" >}}


## `zsd` {#zsd}

`zsd` is a little cli tool to revert a file on the command line.

-   list zfs-snapshots where the given file was modified

<!--listend-->

```sh
main⟩ ./zsd go.mod list
  # | Snapshot                               | Snapshot age
-----------------------------------------------------------
  0 | zfs-auto-snap_hourly-2020-02-12-12h00U | 5 hours
  1 | zfs-auto-snap_hourly-2020-02-12-09h00U | 8 hours
```

-   show the differences between the actual version and from the given snapshot

<!--listend-->

```sh
main⟩ ./zsd go.mod diff 0
Diff from the actual version to the version from: 2020-02-12 10:07:44.434355182 +0100 CET
module github.com/j-keck/zfs-snap-diff

require (
  github.com/BurntSushi/toml v0.3.1
  github.com/j-keck/go-diff v1.0.0
-	github.com/j-keck/plog v0.5.0
+	github.com/j-keck/plog v0.6.0
  github.com/stretchr/testify v1.4.0 // indirect
)

go 1.12
```

-   restore the given file with an older version

<!--listend-->

```sh
main⟩ ./zsd go.mod revert 0
backup from the actual version created at: /home/j/.cache/zfs-snap-diff/backups/home/j/prj/priv/zfs-snap-diff/go.mod_20200212_182709%
```
