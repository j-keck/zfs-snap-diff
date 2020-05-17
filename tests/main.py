import unittest
from zfs import ZFS
import fs

def main() -> None:
    zfs = ZFS()
    zfs.createDataset()

    loader = unittest.TestLoader()
    tests = loader.discover('.')
    testRunner = unittest.runner.TextTestRunner()
    testRunner.run(tests)

    zfs.destroyDataset()


if __name__ == "__main__":
    main()
