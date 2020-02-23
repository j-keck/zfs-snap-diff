+++
title = "Configuration"
draft = false
creator = "Emacs 26.3 (Org mode 9.1.9 + ox-hugo)"
weight = 40
+++

`zfs-snap-diff` loads it's configuration from:

{{< tabs "config-location" >}}
{{< tab "Linux, FreeBSD, Solaris" >}}

```text
$XDG_CONFIG_HOME/.config/zfs-snap-diff/zfs-snap-diff.toml
$HOME/.config/zfs-snap-diff/zfs-snap-diff.toml
```

{{< /tab >}}
{{< tab "macOS" >}}

```text
$HOME/Library/Application Support/zfs-snap-diff/zfs-snap-diff.toml
```

{{< /tab >}}
{{< /tabs >}}

if it does not find a configuration, it will create the following default configuration:

```text
use-cache-dir-for-backups = true
days-to-scan = 2
max-archive-unpacked-size-mb = 200
snapshot-name-template = "zfs-snap-diff-%FT%H:%M"
compare-method = "auto"
diff-context-size = 5

[webserver]
  listen-ip = "127.0.0.1"
  listen-port = 12345
  use-tls = false
  cert-file = ""
  key-file = ""

[zfs]
  use-sudo = false
  mount-snapshots = false
```


## `use-cache-dir-for-backups` {#use-cache-dir-for-backups}

If it's set to `true`, the file backups will be stored in the users cache-directory.

> On Unix systems, it returns $XDG\_CACHE\_HOME as specified by <https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html>
> if non-empty, else $HOME/.cache. On Darwin, it returns $HOME/Library/Caches. On Windows, it returns %LocalAppData%.
> On Plan 9, it returns $home/lib/cache.

<https://golang.org/pkg/os/#UserCacheDir>

If it's `false`, it will create the backup file under the actual directory in the `./zfs-snap-diff/` folder.


## `days-to-scan` {#days-to-scan}

To speedup the scan for other file versions, `zfs-snap-diff` performs the scan incremental
when you request an older file version. This parameter determines how many days are scanned
if you request a older versions.


## `max-archive-unpacked-size-mb` {#max-archive-unpacked-size-mb}

The maximal (unpacked) archive size is restricted by default.
Set this to `-1` to allow disable this restriction.


## `snapshot-name-template` {#snapshot-name-template}

Snapshot name template. Used to create snapshots in the web-app.
The template supports the following format sequences:

```text
Format sequences are alike the `date` command
  %d: day of month (e.g., 01)
  %m: month (01..12)
  %y: last two digits of year (00..99)
  %Y: year
  %F: full date; like %Y-%m-%d
  %H: hour (00..23)
  %I: hour (01..12)
  %M: minute (00..59)
  %S: second (00..60)
  %s: seconds since 1970-01-01 00:00:00 UTC
```


## `compare-method` {#compare-method}

Used compare method to find different file versions.
This is used when scanning the zfs snapshots to determine
if a file was modified in a snapshot.


### auto {#auto}

Uses `md5` for text files and `size+mtime` for others


### size {#size}

If two files versions have the same filesize,
it's interpreted as the same version.


### mtime {#mtime}

If two files versions have the same mtime,
it's interpreted as the same version.


### size+mtime {#size-plus-mtime}

If two files versions have the same size AND mtime,
it's interpreted as the same version.


### content {#content}

If two files versions have the same content,
it's interpreted as the same version.


### md5 {#md5}

If two files versions have the same md5 sum,
it's interpreted as the same version.


## `diff-context-size` {#diff-context-size}

Diff context size in the webui.
