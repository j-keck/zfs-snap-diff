module ZSD.Views.BookmarkManager
       ( getBookmarks
       , addBookmark
       , removeBookmark
       ) where

import Prelude

import Data.Array as A
import Data.HashMap (HashMap)
import Data.HashMap as HM
import Data.Maybe (Maybe, fromMaybe, maybe)
import Effect (Effect)
import ZSD.LocalStorage as LocalStorage
import ZSD.Model.Dataset (Dataset)

type DatasetName = String
type Path = String
type Bookmarks = Array Path
type BookmarksStore = HashMap DatasetName Bookmarks


getBookmarks :: Dataset -> Effect Bookmarks
getBookmarks ds = fromMaybe [] <<< HM.lookup ds.name <$> loadStore

  
addBookmark :: Dataset -> Path -> Effect Bookmarks
addBookmark ds dir = modifyBookmarks ds (_ `A.snoc` dir)


removeBookmark :: Dataset -> Path -> Effect Bookmarks
removeBookmark ds dir = modifyBookmarks ds (A.filter ((/=) dir))


modifyBookmarks :: Dataset -> (Bookmarks -> Bookmarks) -> Effect Bookmarks
modifyBookmarks ds f = do
  store <- loadStore
  let bms = f $ fromMaybe [] $ HM.lookup ds.name store
  saveStore $ HM.insert ds.name bms store
  pure bms


saveStore :: BookmarksStore -> Effect Unit
saveStore store =
  LocalStorage.setItem "bookmarks" $ HM.toArrayBy (\ds bms -> { ds , bms }) store


loadStore :: Effect BookmarksStore
loadStore = do
  store <- fromMaybe (pure []) <$> LocalStorage.getItem "bookmarks"
           :: Effect (Maybe (Array { ds :: String, bms :: Array Path } ))
  pure $ maybe HM.empty (HM.fromArrayBy _.ds _.bms) store
