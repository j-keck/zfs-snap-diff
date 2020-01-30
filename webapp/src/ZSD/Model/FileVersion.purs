module ZSD.Model.FileVersion where

import Prelude

import Affjax.ResponseFormat as ARF
import Control.Alt ((<|>))
import Data.Either (Either(..))
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Maybe (maybe)
import Data.Newtype (unwrap)
import Data.String as S
import Effect.Aff (Aff)
import Simple.JSON (class ReadForeign, readImpl)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError(..))
import ZSD.Model.FSEntry (FSEntry(..))
import ZSD.Model.Snapshot (Snapshot)

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

unwrapFSEntry :: FileVersion -> FSEntry
unwrapFSEntry = case _ of
  ActualVersion file -> file
  BackupVersion { file } -> file


unwrapPath :: FileVersion -> String
unwrapPath = unwrapFSEntry >>> unwrap >>> _.path


uniqueName :: FileVersion -> String
uniqueName = case _ of
  ActualVersion entry -> (unwrap entry).name
  BackupVersion { file: (FSEntry { name }), snapshot } ->
    let { before, after } = maybe { before: name, after: "" }
                                  (flip S.splitAt name)
                                  $ S.lastIndexOf (S.Pattern ".") name
    in before <> "-" <> snapshot.name <> after


isBackupVersion :: FileVersion -> Boolean
isBackupVersion = case _ of
  BackupVersion _ -> true
  _ -> false



restore :: FSEntry -> FileVersion -> Aff (Either AppError String)
restore (FSEntry { path: actualPath }) (BackupVersion { file: (FSEntry {path: backupPath}) }) =
  HTTP.post ARF.string "/api/restore-file" { actualPath, backupPath }
restore _ (ActualVersion _) = pure $ Left $ Bug "restore the actual version not possible"
