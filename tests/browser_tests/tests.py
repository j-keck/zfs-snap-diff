# the test suite - run it per `python -m unittest`
import unittest
from typing import Any
from selenium import webdriver # type: ignore
import sys

import fs
from browser_tests.page import Page
from zfs import ZFS

class Tests(unittest.TestCase):
    zfs = ZFS()

    def setUp(self) -> None:
        self.page = Page(headless = True)
        self.assertIn("ZFS-Snap-Diff", self.page.title())


    def testActualFileContent(self) -> None:
        fs.createTestFile(self.zfs.mountpoint() + "/file.txt",
                             ["firstline", "secondline", "thirdline"]
        )

        self.page.selectView("Browse filesystem")
        self.page.selectDataset(self.zfs.dataset)
        self.page.findByXPath("//td[contains(.,'file.txt')]").click()
        self.assertIn("Current content of file.txt", self.page.findById("file-actions-header").text)
        self.assertIn("firstline\nsecondline\nthirdline", self.page.findByCSS("#file-actions-body > pre > code").text)


    def testCreateSnapshotInBrowseFilesystem(self) -> None:
        self.page.selectView("Browse filesystem")
        self.page.selectDataset(self.zfs.dataset)
        self.page.createSnapshot("create-snapshot-in-browse-filesystem")
        self.assertIn("@create-snapshot-in-browse-filesystem' created", self.page.alertText())


    def testCreateSnapshotInBrowseSnapshots(self) -> None:
        self.page.selectView("Browse snapshots")
        self.page.selectDataset(self.zfs.dataset)
        self.page.createSnapshot("create-snapshot-in-browse-snapshots")
        self.assertIn("@create-snapshot-in-browse-snapshots' created", self.page.alertText())


    def testDestroySnapshot(self) -> None:
        self.page.selectView("Browse snapshots")
        self.page.selectDataset(self.zfs.dataset)


        # create snapshot
        self.page.createSnapshot("destroy-snapshot")
        self.page.closeAlert()

        # destroy snapshot
        self.page.destroySnapshot("destroy-snapshot")
        self.assertIn("Snapshot 'destroy-snapshot' destroyed", self.page.alertText())
        self.page.closeAlert()


    def testRenameSnapshot(self) -> None:
        self.page.selectView("Browse snapshots")
        self.page.selectDataset(self.zfs.dataset)

        # create snapshot
        self.page.createSnapshot("rename-snapshot")
        self.page.closeAlert()

        # rename snapshot
        self.page.renameSnapshot("rename-snapshot", "snapshot-rename")
        self.assertIn("Snapshot 'rename-snapshot' renamed to 'snapshot-rename'", self.page.alertText())
        self.page.closeAlert()


    def testCloneSnapshot(self) -> None:
        self.page.selectView("Browse snapshots")
        self.page.selectDataset(self.zfs.dataset)

        # create snapshot
        self.page.createSnapshot("clone-snapshot")
        self.page.closeAlert()

        # clone snapshot
        self.page.cloneSnapshot("clone-snapshot", "cloned")
        self.assertIn("Snapshot 'clone-snapshot' cloned to '"+self.zfs.pool+"/cloned'", self.page.alertText())
        self.page.closeAlert()


    def testRollbackSnapshot(self) -> None:
        self.page.selectView("Browse snapshots")
        self.page.selectDataset(self.zfs.dataset)

        # create snapshot
        self.page.createSnapshot("rollback-snapshot")
        self.assertIn("@rollback-snapshot' created", self.page.alertText())
        self.page.closeAlert()

        # create a file
        fs.createTestFile(self.zfs.mountpoint() + "/rollback-test.txt", ["dummy"])
        self.assertTrue(fs.exists(self.zfs.mountpoint() + "/rollback-test.txt"))

        # rollback
        self.page.rollbackSnapshot("rollback-snapshot")
        self.assertIn("Snapshot 'rollback-snapshot' rolled back", self.page.alertText())
        self.assertFalse(fs.exists(self.zfs.mountpoint() + "/rollback-test.txt"))


    def tearDown(self) -> None:
        self.page.close()

