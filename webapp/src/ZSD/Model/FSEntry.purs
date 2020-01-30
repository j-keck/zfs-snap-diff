module ZSD.Model.FSEntry where

import Prelude

import Affjax.ResponseFormat as ARF
import Data.Either (Either)
import Data.Newtype (class Newtype, unwrap)
import Data.String as S
import Effect.Aff (Aff)
import Simple.JSON (class ReadForeign, class WriteForeign)
import Web.File.Blob (Blob)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.DateTime (DateTime)
import ZSD.Ops ((<$$>))

newtype FSEntry = FSEntry
  { name    :: String
  , path    :: String
  , kind    :: String
  , size    :: Number
  , modTime :: DateTime
  }

isFile :: FSEntry -> Boolean
isFile = unwrap >>> _.kind >>> (==) "FILE"

isDir :: FSEntry -> Boolean
isDir = unwrap >>> _.kind >>> (==) "DIR"

isLink :: FSEntry -> Boolean
isLink = unwrap >>> _.kind >>> (==) "LINK"

isHidden :: FSEntry -> Boolean
isHidden = unwrap >>> _.name >>> S.take 1 >>> (==) "."

downloadText :: FSEntry -> Aff (Either AppError String)
downloadText (FSEntry { path }) = HTTP.post ARF.string "/api/download" { path }

downloadBlob :: FSEntry -> Aff (Either AppError Blob)
downloadBlob (FSEntry { path }) = HTTP.post ARF.blob "/api/download" { path }

stat :: String -> Aff (Either AppError FSEntry)
stat path = FSEntry <$$> HTTP.post' "/api/stat" { path }



derive newtype instance showFSEntry :: Show FSEntry
derive instance newtypeFSEntry :: Newtype FSEntry _
derive newtype instance readForeignFSEntry :: ReadForeign FSEntry
derive newtype instance writeForeignFSEntry :: WriteForeign FSEntry
instance eqFSEntry :: Eq FSEntry where
  -- entries are equal if their have the same path
  eq (FSEntry { path: p1 }) (FSEntry { path: p2} ) = eq p1 p2
