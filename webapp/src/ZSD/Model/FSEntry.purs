module ZSD.Model.FSEntry where

import Prelude
import Affjax.ResponseFormat as ARF
import Data.Either (Either)
import Data.String as S
import Effect.Aff (Aff)
import Web.File.Blob (Blob)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.DateTime (DateTime)

type FSEntry =
  { name    :: String
  , path    :: String
  , kind    :: String
  , size    :: Number
  , modTime :: DateTime
  }

isFile :: FSEntry -> Boolean
isFile { kind } = kind == "FILE"

isDir :: FSEntry -> Boolean
isDir { kind } = kind == "DIR"

isLink :: FSEntry -> Boolean
isLink { kind } = kind == "LINK"

isHidden :: FSEntry -> Boolean
isHidden { name } = S.take 1 name == "."

downloadText :: FSEntry -> Aff (Either AppError String)
downloadText { path } = HTTP.post ARF.string "/api/download" { path }

downloadBlob :: FSEntry -> Aff (Either AppError Blob)
downloadBlob { path } = HTTP.post ARF.blob "/api/download" { path }
