module ZSD.Model.Dataset where

import ZSD.Model.FSEntry (FSEntry)

type Datasets = Array Dataset

type Dataset =
  { name :: String
  , used :: Number
  , avail :: Number
  , refer :: Number
  , mountPoint :: FSEntry
  }
