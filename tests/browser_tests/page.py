# the web-page actions
from typing import Any, cast
from selenium import webdriver # type: ignore

class Page:

    def __init__(self, headless: bool = True) -> None:
        options = webdriver.FirefoxOptions()
        if headless:
            options.add_argument('--headless')
        self.wd = webdriver.Firefox(options = options)
        self.wd.implicitly_wait(10) # in seconds
        self.wd.get("http://localhost:12345")


    def title(self) -> str:
        return cast(str, self.wd.title)


    def selectView(self, name: str) -> None:
        self.wd.find_element_by_link_text(name).click()


    def selectDataset(self, name: str) -> None:
        path = "//td[contains(.,'{}')]".format(name)
        self.findByXPath(path).click()


    def createSnapshot(self, name: str) -> None:
        self.findById("create-snapshot").click()
        self.findById("snapshot-name-template").clear()
        self.findById("snapshot-name-template").send_keys(name)
        self.findById("confirm-btn-ok").click()


    def destroySnapshot(self, name: str) -> None:
        self.findById("snapshot-actions").click()
        self.findById("destroy-" + name).click()
        self.findById("confirm-btn-ok").click()


    def renameSnapshot(self, name: str, newName: str) -> None:
        self.findById("snapshot-actions").click()
        self.findById("rename-" + name).click()
        self.findById("snapshot-name-template").clear()
        self.findById("snapshot-name-template").send_keys(newName)
        self.findById("confirm-btn-ok").click()


    def cloneSnapshot(self, snapName: str, dsName: str) -> None:
        self.findById("snapshot-actions").click()
        self.findById("clone-" + snapName).click()
        self.findById("fs-name").clear()
        self.findById("fs-name").send_keys(dsName)
        self.findById("confirm-btn-ok").click()


    def rollbackSnapshot(self, name: str) -> None:
        self.findById("snapshot-actions").click()
        self.findById("rollback-" + name).click()
        self.findById("confirm-btn-ok").click()


    def alertText(self) -> str:
        return cast(str, self.findByCSS(".alert").text)


    def closeAlert(self) -> None:
        self.findByCSS(".alert > .close").click()


    def findById(self, id: str) -> Any:
        return self.wd.find_element_by_id(id)


    def findByCSS(self, sel: str) -> Any:
        return self.wd.find_element_by_css_selector(sel)


    def findByXPath(self, path: str) -> Any:
        return self.wd.find_element_by_xpath(path)


    def close(self) -> None:
        self.wd.quit()
