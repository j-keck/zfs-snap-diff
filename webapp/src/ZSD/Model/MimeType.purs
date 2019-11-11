module ZSD.Model.MimeType where

import Data.Either (Either)
import Data.Newtype (class Newtype)
import Data.String.Utils as SU
import Effect.Aff (Aff)
import Prelude (class Eq, class Show, (<$>), (==))
import Simple.JSON (class ReadForeign)
import Simple.JSON as F
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.FSEntry (FSEntry)


newtype MimeType = MimeType String
derive newtype instance showMimeType :: Show MimeType
derive newtype instance eqMimeTime :: Eq MimeType
derive instance newtypeMimeType :: Newtype MimeType _
instance readForeignMimeType :: ReadForeign MimeType where
   readImpl f = toMimeType <$> F.readImpl f
     where toMimeType :: { mimeType :: String } -> MimeType
           toMimeType obj = MimeType obj.mimeType



isText :: MimeType -> Boolean
isText (MimeType mimeType) = SU.startsWith "text/" mimeType

isImage :: MimeType -> Boolean
isImage (MimeType mimeType) = SU.startsWith "image/" mimeType

isPDF :: MimeType -> Boolean
isPDF (MimeType mimeType) = mimeType == "application/pdf"

fetch :: FSEntry -> Aff (Either AppError MimeType)
fetch { path } = HTTP.post' "/api/mime-type" { path }
