module ZSD.Model.FileVersion where

import Data.Either (Either)
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Maybe (maybe)
import Data.String as S
import Effect.Aff (Aff)
import Prelude (class Eq, class Show, flip, ($), (<>))
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Ops ((<$$$>))

type FileVersions = Array FileVersion

data FileVersion =
    ActualVersion FSEntry
  | BackupVersion
    { file :: FSEntry
    , snapshot :: Snapshot
    }

derive instance genericFileVersion :: Generic FileVersion _
derive instance eqFileVersion :: Eq FileVersion
instance showFileVersion :: Show FileVersion where
  show = genericShow

unwrapFile :: FileVersion -> FSEntry
unwrapFile = case _ of
  ActualVersion file -> file
  BackupVersion { file } -> file


uniqueName :: FileVersion -> String
uniqueName = case _ of
  ActualVersion { name } -> name
  BackupVersion { file, snapshot } ->
    let { before, after } = maybe { before: file.name, after: "" }
                                  (flip S.splitAt file.name)
                                  $ S.lastIndexOf (S.Pattern ".") file.name
    in before <> "-" <> snapshot.name <> after



fetch :: FSEntry -> Aff (Either AppError FileVersions)
fetch { path } = BackupVersion <$$$> HTTP.post' "/api/find-file-versions" { path, "compare-method": "auto" }
