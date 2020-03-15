module ZSD.Model.Snapshot where

import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.Utils.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.DateTime (DateTime)
import ZSD.Model.MountPoint (MountPoint)

type Snapshot
  = { name :: String
    , fullName :: String
    , created :: DateTime
    , mountPoint :: MountPoint
    }

type Snapshots
  = Array Snapshot

fetchForDataset :: String -> Aff (Either AppError Snapshots)
fetchForDataset datasetName = HTTP.post' "api/snapshots-for-dataset" { datasetName }
