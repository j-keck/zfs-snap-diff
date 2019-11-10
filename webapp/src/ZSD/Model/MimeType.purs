module ZSD.Model.MimeType where

import Data.Either (Either)
import Data.String.Utils as SU
import Effect.Aff (Aff)
import Prelude ((==))
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.FSEntry (FSEntry)


type MimeType = { mimeType :: String }

isText :: MimeType -> Boolean
isText { mimeType } = SU.startsWith "text/" mimeType

isImage :: MimeType -> Boolean
isImage { mimeType } = SU.startsWith "image/" mimeType

isPDF :: MimeType -> Boolean
isPDF { mimeType } = mimeType == "application/pdf"

fetch :: FSEntry -> Aff (Either AppError MimeType)
fetch { path } = HTTP.post' "/api/mime-type" { path }
