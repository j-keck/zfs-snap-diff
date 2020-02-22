module ZSD.Model.FileVersion where

import Prelude

import Affjax.ResponseFormat as ARF
import Control.Alt ((<|>))
import Data.Either (Either(..))
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Maybe (maybe)
import Data.Newtype (class Newtype, unwrap)
import Data.String as S
import Effect.Aff (Aff)
import Simple.JSON (class ReadForeign, readImpl)

import ZSD.Model.AppError (AppError(..))
import ZSD.Model.DateRange (DateRange)
import ZSD.Model.FH (FH(..))
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Utils.Ops ((<$$>))
import ZSD.Utils.HTTP as HTTP

data FileVersion =
    CurrentVersion FH
  | BackupVersion
    { current   :: FH
    , backup    :: FH
    , snapshot :: Snapshot
    }


uniqueName :: FileVersion -> String
uniqueName = case _ of
  CurrentVersion entry -> (unwrap entry).name
  BackupVersion { backup: (FH { name }), snapshot } ->
    let { before, after } = maybe { before: name, after: "" }
                                  (flip S.splitAt name)
                                  $ S.lastIndexOf (S.Pattern ".") name
    in before <> "-" <> snapshot.name <> after




scanBackups :: FH -> DateRange -> Aff (Either AppError ScanResult)
scanBackups e dateRange
  = ScanResult <$$> HTTP.post' "api/find-file-versions" { path: (unwrap >>> _.path) e, dateRange }


newtype ScanResult = ScanResult
  { fileVersions          :: Array FileVersion
  , dateRange             :: DateRange
  , snapsScanned          :: Int
  , snapsToScan           :: Int
  , snapsFileMissing      :: Int
  , lastScannedSnapshot   :: Snapshot
  , scanDuration          :: Number
  }


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




derive instance genericFileVersion :: Generic FileVersion _
derive instance eqFileVersion :: Eq FileVersion
instance showFileVersion :: Show FileVersion where
  show = genericShow
instance readForeignFileVersion :: ReadForeign FileVersion where
  readImpl f =     BackupVersion <$> readImpl f
               <|> CurrentVersion <$> readImpl f





restore :: FileVersion -> Aff (Either AppError String)
restore (BackupVersion { current, backup } ) =
  let currentPath = (unwrap >>> _.path) current
      backupPath = (unwrap >>> _.path) backup
  in HTTP.post ARF.string "api/restore-file" { currentPath, backupPath }
restore (CurrentVersion _) = pure $ Left $ Bug "restore the current version not possible"
