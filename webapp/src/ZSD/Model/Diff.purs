module ZSD.Model.Diff where

import Prelude

import Affjax.ResponseFormat as ARF
import Data.Either (Either(..))
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Newtype (unwrap)
import Effect.Aff (Aff)
import Foreign (ForeignError(..))
import Foreign as Foreign
import Simple.JSON (class ReadForeign)

import ZSD.Utils.HTTP as HTTP
import ZSD.Model.AppError (AppError(..))
import ZSD.Model.FileVersion (FileVersion(..))


type Diff =
  { deltas :: Array Deltas
  , patches :: Array String
  , sideBySideDiffHTMLFragment :: Array String
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


fetch :: FileVersion -> Aff (Either AppError Diff)
fetch (BackupVersion { actual, backup }) =
  let actualPath = (unwrap >>> _.path) actual
      backupPath = (unwrap >>> _.path) backup
  in HTTP.post' "/api/diff" { actualPath, backupPath }
fetch (ActualVersion _ ) = pure $ Left $ Bug "diff with the same version not possible"



revert :: FileVersion -> Int -> Aff (Either AppError String)
revert (BackupVersion { actual, backup}) deltaIdx =
  let actualPath = (unwrap >>> _.path) actual
      backupPath = (unwrap >>> _.path) backup
  in HTTP.post ARF.string "/api/revert-change" { actualPath, backupPath, deltaIdx }
revert (ActualVersion _) _ = pure $ Left $ Bug "revert for the actual version not possible"
