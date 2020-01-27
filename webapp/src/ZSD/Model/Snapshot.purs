module ZSD.Model.Snapshot where

import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.DateTime (DateTime)
import ZSD.Model.FSEntry (FSEntry)

type Snapshot =
  { name     :: String
  , fullName :: String
  , created  :: DateTime
  , dir      :: FSEntry
  }


type Snapshots = Array Snapshot


fetchForDataset :: Dataset -> Aff (Either AppError Snapshots)
fetchForDataset { name } = HTTP.post' "/api/snapshots-for-dataset" { datasetName: name }
