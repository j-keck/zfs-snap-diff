from typing import Any, Union, List
import subprocess
import os


class ZFS:
    pool: str
    dataset: str

    def __init__(self, dataset: str = "zsd-e2e-test") -> None:
        self.pool = self.lookupPool()
        self.dataset = dataset


    def mountpoint(self) -> str:
        return "/tmp/{}".format(self.dataset)


    def createDataset(self) -> None:
        args = "create -o mountpoint={} {}/{}".format(self.mountpoint(), self.pool, self.dataset)
        self.zfs(args)

        # fix permission
        username = os.getlogin()
        subprocess.run(["sudo", "chown", username, self.mountpoint()])


    def destroyDataset(self) -> None:
        self.zfs("destroy -R {}/{}".format(self.pool, self.dataset))


    @classmethod
    def zfs(cls, args: Union[str, List[str]]) -> None:
        if isinstance(args, str):
            args = args.split()
        subprocess.run(["sudo", "zfs"] + args)


    @classmethod
    def lookupPool(cls) -> str:
        # dirty, unsafe code!
        out = subprocess.run(["sudo", "zpool", "list", "-Ho", "name"]
                             , text=True
                             , capture_output=True).stdout
        return out.strip()

