module ZSD.Model.MountPoint
       ( MountPoint(..)
       , MountPointFields
       ) where

import Prelude

import Data.Newtype (class Newtype)
import Simple.JSON (class ReadForeign)
import Simple.JSON as F
import ZSD.Model.DateTime (DateTime)

newtype MountPoint = MountPoint MountPointFields

type MountPointFields =
  { name  :: String
  , path  :: String
  , size  :: Number
  , mtime :: DateTime
  }

derive newtype instance showMountPoint :: Show MountPoint
derive newtype instance eqMountPoint :: Eq MountPoint
derive instance newtypeMountPoint :: Newtype MountPoint _
instance readForeignMountPoint :: ReadForeign MountPoint where
  readImpl f = MountPoint <$> F.readImpl f
