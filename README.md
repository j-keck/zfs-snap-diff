# Background
  
I make every 5 minutes a snapshot (keep it for 1 day) and once a day for long term (keep it for one month) from my home partition on a ZFS filesystem.
If i messed up a file, i need to search a clean state from the file in the snapshots - not always easy if don't realize it directly.

`zfs-snap-diff` is a little tool to help me for such cases.


# Description

With `zfs-snap-diff` you can explore file differences from different zfs snapshots.

  
`zfs-snap-diff` has a web frontend, so it can run on your local work machine or on your remote file / backup server (no Xserver necesarry).

To keep it protable and independent, it's made as a single executable with all html / js stuff included.
The backend is implemented in golang, the frontend with [angularjs](https://angularjs.org), [bootstrap](http://getbootstrap.com) and [jsdifflib](https://github.com/cemerick/jsdifflib).


  
*!! it's in a very early dev state - only tested on FreeBSD !!*



#Usage


### Startup a server instance

      ./zfs_snap_diff <ZFS_NAME>

### Connect with your web browser

      http://localhost:12345

### Search a file
  
Search a file in the file browser.
    
![File browser](doc/zsd-file-browser.png)

### Select a file

When a file is selected, `zsd-snap-diff` search all snapshots where the selected file was modified (currently it compares only mod-time and file-size).
    
![File selected](doc/zsd-file-selected.png)
  


### Select a snapshot


When you select a snapshot, and

  * the file size is > 100MB: it downloads the selected file
  * the file is a text file: it shows a diff from the selected snapshot to the current state
  * the file is a binary file: it embed the file (per embed tag)

![File Diff](doc/zsd-snap-selected.png)  



#Notes

  * if you download a file from a snapshot, the generated file name has the snapshot name included:

        <ORG_FILE_NAME>-<SNAPSHOT_NAME>.<FILE_SUFFIX>
  
  * for snapshot differences, you need to set the diff permission:

        zfs allow -u <USER_NAME> diff <ZFS_NAME>


  


  
# Build:

  * clone the repository

        git clone github.com/j-keck/zfs-snap-diff

  * change into the project directory

        cd zfs-snap-diff

  * init submodule

        git submodule init

  * update submodule

        git submodule update

  * generate golang src from static web content (this generates bindata.go)
  
        go-bindata webapp/...

  * build it
  
        go build -ldflags "-X main.VERSION $(git describe)"


  
# Run:
  
        ./zfs-snap-diff <ZFS_NAME> 

  * starts a web server on port http://127.0.0.1:12345
  * check `-h` for currently supported parameters


### for dev:
  
        ZSD_SERVE_FROM_WEBAPP=YES ./zfs-snap-diff <ZFS_NAME> 

  * serve static content from webapp dir


  
# Changelog

###0.0.X###
0.0.2 :
  * partial frontend configuration from server
  * fix firefox ui


0.0.1:
  * prototype  
