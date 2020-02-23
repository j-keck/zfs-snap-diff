module ZSD.Model.Dataset where

import Affjax.ResponseFormat as ARF
import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.Model.AppError (AppError)
import ZSD.Model.MountPoint (MountPoint)
import ZSD.Utils.HTTP as HTTP

type Datasets
  = Array Dataset

type Dataset
  = { name :: String
    , used :: Number
    , avail :: Number
    , refer :: Number
    , mountPoint :: MountPoint
    }

createSnapshot :: Dataset -> String -> Aff (Either AppError String)
createSnapshot ds name =
  HTTP.post ARF.string "api/create-snapshot"
    { datasetName: ds.name
    , snapshotName: name
    }

destroySnapshot :: Dataset -> String -> Aff (Either AppError String)
destroySnapshot ds name =
  HTTP.post ARF.string "api/destroy-snapshot"
    { datasetName: ds.name
    , snapshotName: name
    }
