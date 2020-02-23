module ZSD.Utils.BookmarkManager
  ( get
  , add
  , remove
  , contains
  ) where

import Prelude
import Data.Array as A
import Data.HashMap (HashMap)
import Data.HashMap as HM
import Data.Maybe (Maybe, fromMaybe, isJust, maybe)
import Data.Newtype (unwrap)
import Effect (Effect)
import Effect.Unsafe (unsafePerformEffect)
import ZSD.Utils.LocalStorage as LocalStorage
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FH (FH)

type DatasetName
  = String

type Bookmarks
  = Array FH

type BookmarksStore
  = HashMap DatasetName Bookmarks

get :: Dataset -> Effect Bookmarks
get ds = fromMaybe [] <<< HM.lookup ds.name <$> loadStore

add :: Dataset -> FH -> Effect Bookmarks
add ds dir = modify ds (_ `A.snoc` dir)

contains :: Dataset -> FH -> Boolean
contains ds dir =
  let
    p = (unwrap >>> _.path) dir
  in
    unsafePerformEffect $ A.find (\e -> (unwrap >>> _.path $ e) == p) >>> isJust <$> get ds

remove :: Dataset -> FH -> Effect Bookmarks
remove ds dir = modify ds (A.filter ((/=) dir))

modify :: Dataset -> (Bookmarks -> Bookmarks) -> Effect Bookmarks
modify ds f = do
  store <- loadStore
  let
    bms = f $ fromMaybe [] $ HM.lookup ds.name store
  saveStore $ HM.insert ds.name bms store
  pure bms

saveStore :: BookmarksStore -> Effect Unit
saveStore store = LocalStorage.setItem "bookmarks" $ HM.toArrayBy (\ds bms -> { ds, bms }) store

loadStore :: Effect BookmarksStore
loadStore = do
  store <-
    fromMaybe (pure []) <$> LocalStorage.getItem "bookmarks" ::
      Effect (Maybe (Array { ds :: String, bms :: Bookmarks }))
  pure $ maybe HM.empty (HM.fromArrayBy _.ds _.bms) store
