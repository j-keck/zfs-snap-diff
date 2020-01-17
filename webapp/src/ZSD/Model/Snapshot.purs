module ZSD.Model.Snapshot where

import ZSD.Model.DateTime (DateTime)

type Snapshot =
  { name :: String
  , created :: DateTime
  }
