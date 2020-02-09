module ZSD.Model.FH where

import Prelude

import Affjax.ResponseFormat as ARF
import Data.Either (Either)
import Data.Generic.Rep (class Generic)
import Data.Generic.Rep.Show (genericShow)
import Data.Maybe (maybe)
import Data.Newtype (class Newtype, over, unwrap)
import Data.String as S
import Data.Traversable as T
import Effect.Aff (Aff)
import Record as Record
import Simple.JSON (class ReadForeign, class WriteForeign)
import Web.File.Blob (Blob)
import ZSD.Model.AppError (AppError)
import ZSD.Model.DateTime (DateTime)
import ZSD.Model.Kind (Kind(..))
import ZSD.Model.MimeType (MimeType)
import ZSD.Model.MountPoint (MountPoint(..))
import ZSD.Utils.HTTP as HTTP
import ZSD.Utils.Ops ((<$$>), (</>))


newtype FH = FH
  { name     :: String
  , path     :: String
  , kind     :: Kind
  , size     :: Number
  , mtime    :: DateTime
  }


fromMountPoint :: MountPoint -> FH
fromMountPoint (MountPoint rec) =
  FH $ Record.merge rec {kind: Dir }


ls :: FH -> Aff (Either AppError (Array FH))
ls dir = HTTP.post' "/api/dir-listing" { path: (unwrap >>> _.path) dir }


stat :: String -> Aff (Either AppError FH)
stat path = FH <$$> HTTP.post' "/api/stat" { path }

stat' :: Array String -> Aff (Either AppError (Array FH))
stat' = map T.sequence <<< T.traverse stat



fetchMimeType :: FH -> Aff (Either AppError MimeType)
fetchMimeType e = HTTP.post' "/api/mime-type" { path: (unwrap >>> _.path) e }


downloadText :: FH -> Aff (Either AppError String)
downloadText e = HTTP.post ARF.string "/api/download" { path: (unwrap >>> _.path) e }


downloadBlob :: FH -> Aff (Either AppError Blob)
downloadBlob e = HTTP.post ARF.blob "/api/download" { path: (unwrap >>> _.path) e }



newtype From = From MountPoint
derive instance newtypeFrom :: Newtype From _
derive instance genericFrom :: Generic From _
instance showFrom :: Show From where
  show = genericShow

newtype To = To MountPoint
derive instance newtypeTo :: Newtype To _
derive instance genericTo :: Generic To _
instance showTo :: Show To where
  show = genericShow

switchMountPoint :: From -> To -> FH -> FH
switchMountPoint (From (MountPoint from)) (To (MountPoint to)) =
  over FH (\fh -> fh { path = switch fh.path })
  where switch p = maybe to.path (to.path </> _) $
                             S.stripPrefix (S.Pattern from.path) p
                         >>= S.stripPrefix (S.Pattern "/")




derive newtype instance showFH :: Show FH
derive instance newtypeFH :: Newtype FH _
derive newtype instance readForeignFH :: ReadForeign FH
derive newtype instance writeForeignFH :: WriteForeign FH
instance eqFH :: Eq FH where
  -- fs entries are equal if their have the same path
  eq (FH { path: p1 }) (FH { path: p2 } ) =
    eq p1 p2
