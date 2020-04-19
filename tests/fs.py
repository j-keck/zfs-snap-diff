# simple file-system utils
from typing import List
import os

def createTestFile(path: str, lines: List[str]) -> None:
    with open(path, "w") as fh:
        for line in lines:
            fh.write(line + "\n")


def exists(path: str) -> bool:
    return os.path.isfile(path)
