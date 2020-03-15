module ZSD.Model.Dataset where

import Affjax.ResponseFormat as ARF
import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.Model.AppError (AppError)
import ZSD.Model.MountPoint (MountPoint)
import ZSD.Model.Snapshot (Snapshot)
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

-- | fetches all datasets from the server
fetch :: Aff (Either AppError Datasets)
fetch = HTTP.get' "api/rescan-datasets"

createSnapshot :: Dataset -> String -> Aff (Either AppError String)
createSnapshot ds name =
  HTTP.post ARF.string "api/create-snapshot"
    { datasetName: ds.name
    , snapshotName: name
    }

destroySnapshot :: Dataset -> Snapshot -> Array String -> Aff (Either AppError String)
destroySnapshot ds snap destroyFlags =
  HTTP.post ARF.string "api/destroy-snapshot"
    { datasetName: ds.name
    , snapshotName: snap.name
    , destroyFlags
    }

renameSnapshot :: Dataset -> Snapshot -> String -> Aff (Either AppError String)
renameSnapshot ds snap newSnapshotName =
  HTTP.post ARF.string "api/rename-snapshot"
    { datasetName: ds.name
    , oldSnapshotName: snap.name
    , newSnapshotName
    }

cloneSnapshot :: Dataset -> Snapshot -> Array String -> String -> Aff (Either AppError String)
cloneSnapshot ds snap cloneFlags fsName =
  HTTP.post ARF.string "api/clone-snapshot"
    { datasetName: ds.name
    , snapshotName: snap.name
    , cloneFlags
    , fsName
    }

rollbackSnapshot :: Dataset -> Snapshot -> Array String -> Aff (Either AppError String)
rollbackSnapshot ds snap rollbackFlags =
  HTTP.post ARF.string "api/rollback-snapshot"
    { datasetName: ds.name
    , snapshotName: snap.name
    , rollbackFlags
    }
