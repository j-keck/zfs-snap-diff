module ZSD.Model.Snapshot where

import Data.Either (Either)
import Effect.Aff (Aff)

import ZSD.Utils.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.DateTime (DateTime)
import ZSD.Model.MountPoint (MountPoint)

type Snapshot =
  { name       :: String
  , fullName   :: String
  , created    :: DateTime
  , mountPoint :: MountPoint
  }


type Snapshots = Array Snapshot


fetchForDataset :: Dataset -> Aff (Either AppError Snapshots)
fetchForDataset { name } = HTTP.post' "api/snapshots-for-dataset" { datasetName: name }
