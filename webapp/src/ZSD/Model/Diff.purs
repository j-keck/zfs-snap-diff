module ZSD.Model.Diff where

import Prelude

import Data.Either (Either(..))
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Newtype (class Newtype)
import Effect.Aff (Aff)
import Foreign (ForeignError(..))
import Foreign as Foreign
import Simple.JSON (class ReadForeign)
import Simple.JSON as F
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError(..))
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..))


type Diff =
  { deltas :: Array Deltas
  , patches :: Array String
  , sideBySideHTMLFragment :: Array String
  , inlineDiffHTMLFragment :: Array String
  }


data DeltaKind =
    Del
  | Eq
  | Ins

derive instance genericDeltaKind :: Generic DeltaKind _
derive instance eqDeltaKind :: Eq DeltaKind
instance showDeltaKind :: Show DeltaKind where
  show = genericShow
instance readForeignDeltaKind :: ReadForeign DeltaKind where
  readImpl f = Foreign.readInt f >>= case _  of
                 (-1) -> pure Del
                 0 -> pure Eq
                 1 -> pure Ins
                 _ -> Foreign.fail (ForeignError "invalid 'DeltaKind'")


type Delta =
  { kind           :: DeltaKind
  , lineNrFrom     :: Int
  , lineNrTarget   :: Int
  , startPosFrom   :: Int
  , startPosTarget :: Int
  , text           :: String
  }


type Deltas = Array Delta


fetch :: FSEntry -> FileVersion -> Aff (Either AppError Diff)
fetch { path } (BackupVersion { file }) = HTTP.post' "/api/diff" { "actual-path": path, "backup-path": file.path}
fetch _ (ActualVersion _ ) = pure $ Left $ Bug "diff with the same version not possible"
