from typing import Any, Union, List
import subprocess
import os

def createDataset(pool: str, name: str) -> str:
    mountpoint = "/tmp/{}".format(name)
    args = "create -o mountpoint={} {}/{}".format(mountpoint, pool, name)
    zfs(args)

    # fix permission
    username = os.getlogin()
    subprocess.run(["sudo", "chown", username, mountpoint])

    return mountpoint


def destroyDataset(pool: str, name: str) -> None:
    zfs("destroy -R {}/{}".format(pool, name))


def zfs(args: Union[str, List[str]]) -> None:
    if isinstance(args, str):
        args = args.split()
    subprocess.run(["sudo", "zfs"] + args)

