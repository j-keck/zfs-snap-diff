module ZSD.Model.Snapshot where

import ZSD.Model.DateTime (DateTime)
import ZSD.Model.FSEntry (FSEntry)

type Snapshot =
  { name    :: String
  , created :: DateTime
  , dir     :: FSEntry
  }

