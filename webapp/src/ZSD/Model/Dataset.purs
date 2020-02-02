module ZSD.Model.Dataset where

import ZSD.Model.MountPoint (MountPoint)

type Datasets = Array Dataset

type Dataset =
  { name       :: String
  , used       :: Number
  , avail      :: Number
  , refer      :: Number
  , mountPoint :: MountPoint
  }
