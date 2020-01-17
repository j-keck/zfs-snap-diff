module ZSD.Model.FileVersion where

import Prelude
import Data.Either (Either(..))
import Control.Alt ((<|>))
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Maybe (maybe)
import Effect.Aff (Aff)
import Data.String as S
import Simple.JSON (class ReadForeign, readImpl)
import ZSD.HTTP as HTTP
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Model.AppError (AppError(..))

type FileVersions = Array FileVersion

-- FIXME: include the FSEntry from the actual version in the BackupVersion and
-- adjust all functions to receive only the FileVersion
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
instance readForeignFileVersion :: ReadForeign FileVersion where
  readImpl f =     BackupVersion <$> readImpl f
               <|> ActualVersion <$> readImpl f

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


isBackupVersion :: FileVersion -> Boolean
isBackupVersion = case _ of
  BackupVersion _ -> true
  _ -> false



restore :: FSEntry -> FileVersion -> Aff (Either AppError Unit)
restore { path } (BackupVersion { file }) = HTTP.post_ "/api/restore-file"
                                             { "actualPath": path, "backupPath": file.path }
restore _ (ActualVersion _) = pure $ Left $ Bug "restore the actual version not possible"
