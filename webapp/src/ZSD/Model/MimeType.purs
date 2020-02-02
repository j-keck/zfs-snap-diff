module ZSD.Model.MimeType where

import Data.Newtype (class Newtype)
import Data.String.Utils as SU
import Prelude (class Eq, class Show, (<$>), (==))
import Simple.JSON (class ReadForeign)
import Simple.JSON as F

newtype MimeType = MimeType String


isText :: MimeType -> Boolean
isText (MimeType mimeType) = SU.startsWith "text/" mimeType


isImage :: MimeType -> Boolean
isImage (MimeType mimeType) = SU.startsWith "image/" mimeType


isPDF :: MimeType -> Boolean
isPDF (MimeType mimeType) = mimeType == "application/pdf"


derive newtype instance showMimeType :: Show MimeType
derive newtype instance eqMimeTime :: Eq MimeType
derive instance newtypeMimeType :: Newtype MimeType _
instance readForeignMimeType :: ReadForeign MimeType where
   readImpl f = toMimeType <$> F.readImpl f
     where toMimeType :: { mimeType :: String } -> MimeType
           toMimeType obj = MimeType obj.mimeType

