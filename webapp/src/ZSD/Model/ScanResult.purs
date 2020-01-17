module ZSD.Model.ScanResult where

import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.DateRange (DateRange)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersions)
import ZSD.Model.Snapshot (Snapshot)


type ScanResult =
  { fileVersions            :: FileVersions
  , scannedDateRange        :: DateRange
  , skippedSnapshots        :: Int
  , scannedSnapshots        :: Int
  , snapshotsToScan         :: Int
  , fileMissingSnapshots    :: Int
  , lastScannedSnapshot     :: Snapshot
  , scanDuration            :: Number
  }

fetch :: FSEntry -> DateRange -> Aff (Either AppError ScanResult)
fetch { path } dateRange = HTTP.post' "/api/find-file-versions" { path, dateRange }

