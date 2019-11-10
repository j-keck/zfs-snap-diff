module ZSD.Model.Dataset where

import ZSD.Model.DateTime

type Datasets = Array Dataset

type Dataset =
  { name :: String
  , used :: Number
  , avail :: Number
  , refer :: Number
  , mountPoint :: MountPoint
  }

type MountPoint =
  { name :: String
  , path :: String
  , kind :: String
  , size :: Number
  , modTime :: DateTime
  }

