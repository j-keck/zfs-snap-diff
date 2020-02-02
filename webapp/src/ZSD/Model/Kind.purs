module ZSD.Model.Kind where

import Prelude

import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Foreign (ForeignError(..))
import Foreign as Foreign
import Simple.JSON (class ReadForeign, class WriteForeign)
import Simple.JSON as F


data Kind =
    File
  | Dir
  | Link
  | Pipe
  | Socket
  | Dev

icon :: Kind -> String
icon = case _ of
  File -> "fas fa-file p-1"
  Dir -> "fas fa-folder p-1"
  Link -> "fas fa-link p-1"
  _ -> "fas fa-hdd p-1"



derive instance genericKind :: Generic Kind _
instance showKind :: Show Kind where
  show = genericShow
instance readForeignKind :: ReadForeign Kind where
  readImpl f = F.readImpl f >>= case _ of
    "FILE" -> pure File
    "DIR" -> pure Dir
    "LINK" -> pure Link
    "PIPE" -> pure Pipe
    "SOCKET" -> pure Socket
    "DEV" -> pure Dev
    s -> Foreign.fail (ForeignError $ "Invalid Kind: '" <> s <> "'")
instance writeForeignKind :: WriteForeign Kind where
  writeImpl kind = F.writeImpl $ case kind of
    File -> "FILE"
    Dir -> "DIR"
    Link -> "LINK"
    Pipe -> "PIPE"
    Socket -> "SOCKET"
    Dev -> "DEV"
         
