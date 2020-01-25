module ZSD.Model.ScanResult where

import Data.Either (Either)
import Data.Newtype (class Newtype)
import Data.Semigroup ((<>))
import Effect.Aff (Aff)
import Prelude (class Eq, class Semigroup, class Show, (+))
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.DateRange (DateRange)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersions)
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Ops ((<$$>))


newtype ScanResult = ScanResult
  { fileVersions          :: FileVersions
  , dateRange             :: DateRange
  , snapsScanned          :: Int
  , snapsToScan           :: Int
  , snapsFileMissing      :: Int
  , lastScannedSnapshot   :: Snapshot
  , scanDuration          :: Number
  }
 
fetch :: FSEntry -> DateRange -> Aff (Either AppError ScanResult)
fetch { path } dateRange = ScanResult <$$> HTTP.post' "/api/find-file-versions" { path, dateRange }



derive newtype instance showScanResult :: Show ScanResult
derive newtype instance eqScanResult :: Eq ScanResult
derive instance newtypeScanResult :: Newtype ScanResult _

instance semigroupScanResult :: Semigroup ScanResult where
  append (ScanResult a) (ScanResult b) =
    ScanResult { fileVersions: a.fileVersions <> b.fileVersions
               , dateRange: a.dateRange <> b.dateRange
               , snapsScanned: a.snapsScanned + b.snapsScanned
               , snapsToScan: b.snapsToScan
               , snapsFileMissing: b.snapsFileMissing
               , lastScannedSnapshot: b.lastScannedSnapshot
               , scanDuration: a.scanDuration + b.scanDuration
               }

