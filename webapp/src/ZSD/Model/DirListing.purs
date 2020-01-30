module ZSD.Model.DirListing where

import Prelude
import Data.Array as A
import Data.Either (Either)
import Effect.Aff (Aff)
import ZSD.HTTP as HTTP
import ZSD.Model.AppError (AppError)
import ZSD.Model.FSEntry (FSEntry(..))
import ZSD.Model.FSEntry as FSEntry

type DirListing = Array FSEntry

filter :: forall r. { showHidden :: Boolean | r } -> DirListing -> DirListing
filter p = A.filter (\e -> (FSEntry.isHidden e) == p.showHidden)


fetch :: FSEntry -> Aff (Either AppError DirListing)
fetch (FSEntry { path }) = HTTP.post' "/api/dir-listing" { path }
