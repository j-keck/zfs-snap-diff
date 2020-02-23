module ZSD.Fragments.DirBrowser where

import Prelude

import Data.Array as A
import Data.Either (Either(..), either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), maybe)
import Data.Monoid (guard)
import Data.Newtype (unwrap)
import Data.String as S
import Data.Traversable as T
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Unsafe (unsafePerformEffect)
import React.Basic (Component, JSX, createComponent, make, fragment)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import Unsafe.Coerce (unsafeCoerce)
import Web.HTML (window)
import Web.HTML.Location (assign)
import Web.HTML.Window (location)
import ZSD.Components.DropDownButton (dropDownButton)
import ZSD.Components.Scroll as Scroll
import ZSD.Components.Spinner as Spinner
import ZSD.Components.Table (table)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FH (FH(..), From(..), To(..), switchMountPoint)
import ZSD.Model.FH as FH
import ZSD.Model.Kind (Kind(..), icon)
import ZSD.Model.MountPoint (MountPoint)
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Utils.BookmarkManager as BM
import ZSD.Utils.Formatter as Formatter
import ZSD.Utils.Ops (tupleM, (</>))
import ZSD.Views.Messages as Messages


type Props =
  { ds              :: Dataset
  , snapshot        :: Maybe Snapshot
  , onFileSelected  :: FH -> Effect Unit
  , onDirSelected   :: FH -> Effect Unit
  }

type Path = String

type State =
  { breadcrumb     :: Array FH
  , dirListing     :: Array FH
  , currentDir     :: FH
  , selectedFile   :: Maybe FH
  , showBrowser    :: Boolean
  , showHidden     :: Boolean
  }

type Self = React.Self Props State

data Command =
    StartAt MountPoint
  | PickFromBreadcrumb FH
  | PickFromBookmark FH
  | OnClick FH
  | ReadDir FH
  | SwitchRoot From To
  | DownloadArchive

update :: Self -> Command -> Effect Unit
update self = case _ of

  StartAt mp -> do
    let fh = FH.fromMountPoint mp
    update self $ ReadDir fh
    bms <- BM.get self.props.ds
    self.setState _ { breadcrumb = [fh], selectedFile = Nothing }



  PickFromBreadcrumb fh ->
       update self (ReadDir fh)
    *> self.setState _ { selectedFile = Nothing, showBrowser = true }
    *> self.props.onDirSelected fh
    *> rebuildBreadcrumb self fh


  PickFromBookmark fh ->
    let fh' = maybe fh
                    (\snap -> switchMountPoint (From self.props.ds.mountPoint) (To snap.mountPoint) fh)
                    self.props.snapshot
    in    update self (ReadDir fh')
       *> self.setState _ { selectedFile = Nothing, showBrowser = true }
       *> self.props.onDirSelected fh
       *> rebuildBreadcrumb self fh'


  ReadDir fh ->
       Spinner.display
    *> self.setState _ { currentDir = fh }
    *> launchAff_ (    FH.ls fh
                   >>= either Messages.appError (\ls -> self.setState _ { dirListing = ls, currentDir = fh } *> Spinner.remove)
                   >>> liftEffect)



  OnClick fh -> Scroll.scrollToTop *> do
    case (unwrap fh).kind of

      Dir -> do
        update self $ ReadDir fh
        self.setState \s -> s { selectedFile = Nothing
                              , breadcrumb = s.breadcrumb `A.snoc` fh
                              }
        self.props.onDirSelected fh

      File -> do
        self.setState _ { showBrowser = false, selectedFile = Just fh }
        self.props.onFileSelected fh
      _ -> pure unit


  SwitchRoot old new -> Spinner.display *> do
    launchAff_ $ do
      let fh = switchMountPoint old new self.state.currentDir
      res <- FH.ls fh
      liftEffect $ case res of
        Right ls ->    self.setState _ { dirListing = ls, currentDir = fh }
                    *> rebuildBreadcrumb self fh
                    *> Spinner.remove
        Left err ->    (Messages.error
                           $ "Directory " <> (unwrap >>> _.name) fh
                          <> " does not exist in the snapshot " <> (unwrap >>> unwrap >>> _.name $ new))
                    *> update self (StartAt $ unwrap new)


  DownloadArchive ->
       Spinner.display
    *> launchAff_ (    FH.prepareArchive self.state.currentDir self.props.snapshot
                   >>= either Messages.appError
                              (\name -> Spinner.remove *> window >>= location >>= assign ("api/download-archive?name=" <> name))
                   >>> liftEffect)




dirBrowser :: Props -> JSX
dirBrowser = make component { initialState, render, didMount, didUpdate }

  where

    component :: Component Props
    component  = createComponent "DirBrowser"


    initialState = { breadcrumb: [], dirListing: []
                   , currentDir: FH {name: "", path: "", kind: Dir, size: 0.0, mtime: bottom}
                   , selectedFile: Nothing, showBrowser: true
                   , showHidden: false
                   }


    didMount self = update self (StartAt $ maybe self.props.ds.mountPoint _.mountPoint self.props.snapshot)


    didUpdate self {prevProps} = do
      guard (self.props.ds /= prevProps.ds) $
        update self (StartAt $ maybe self.props.ds.mountPoint _.mountPoint self.props.snapshot)

      foldMap (\(Tuple old new) ->
                guard (old /= new) $ update self (SwitchRoot (From old.mountPoint) (To new.mountPoint)))
              (tupleM prevProps.snapshot self.props.snapshot)


    render self =
      R.div
      { className: "mt-3"
      , children:
        [ breadcrumb self
        , guard self.state.showBrowser $
             table
              { header: [ R.text "Name", R.text "Size"
                        , fragment [ R.text "Modify time"
                                   , R.span
                                     { className: "fas fa-file-archive float-right p-1 pointer"
                                     , title: "Download the current directory as zip-archive"
                                     , onClick: capture_ $ update self DownloadArchive
                                     }
                                   ]
                        ]
              , rows: filter self.state self.state.dirListing
              , mkRow: \f@(FH { name, size, mtime }) ->
                         [ R.span { className: (unwrap >>> _.kind >>> icon) f } <> R.text name
                         , R.text $ Formatter.filesize size
                         , R.text $ Formatter.dateTime mtime
                         ]
              , onRowSelected: update self <<< OnClick
              }
        ]
      }



    breadcrumb self =
      R.nav_
       [ R.ol
          { className: "breadcrumb"
          , children:
            A.concat
            [ map (\h -> R.li
                         { className: "breadcrumb-item"
                         , children:
                           [ R.a
                             { onClick: capture_ $ update self (PickFromBreadcrumb h)
                             , href: "#"
                             , children: [ R.text (unwrap h).name ]
                             }
                           ]
                         }
                  ) self.state.breadcrumb
              `A.snoc` (maybe mempty (\f -> R.li
                                            { className: "breadcrumb-item"
                                            , children: [ R.text (unwrap f).name ]
                                            })
                        $ self.state.selectedFile)
            , [ R.div
                { className: "ml-auto"
                , children:
                  [ let dir = maybe self.state.currentDir
                                    (\snap -> switchMountPoint (From snap.mountPoint) (To self.props.ds.mountPoint) self.state.currentDir)
                                    self.props.snapshot
                    in dropDownButton
                    { content: let fa = if BM.contains self.props.ds dir then "fas" else "far"
                               in R.span { className: fa <> " fa-bookmark px-2" }
                    , title: "Save bookmark"
                    , disabled: false
                    , onClick: (if BM.contains self.props.ds dir
                                  then BM.remove self.props.ds dir
                                  else BM.add self.props.ds dir) *> self.setState identity
                    , entries: map (\bm -> Tuple (R.text $ (unwrap >>> _.path) bm) (update self $ PickFromBookmark bm))
                                    (unsafePerformEffect $ BM.get self.props.ds)
                    , entriesTitle: "Saved bookmarks"
                    }
                  ]
                }
              ]
            ]
          }
       ]






filter :: forall r. { showHidden :: Boolean | r } -> Array FH -> Array FH
filter p = A.filter (\e -> isHidden e == p.showHidden)
  where isHidden = unwrap >>> _.name >>> S.take 1 >>> (==) "."







null :: forall a. a
null = unsafeCoerce {}



rebuildBreadcrumb :: Self -> FH -> Effect Unit
rebuildBreadcrumb self fh = launchAff_ do
  let root = maybe self.props.ds.mountPoint _.mountPoint self.props.snapshot
      pathElements = unwrap >>> _.path >>> S.split (S.Pattern "/") >>> T.scanl (</>) "" $ fh
      p = A.dropWhile ((/=) (unwrap >>>_.path $ root)) pathElements
  FH.stat' p >>= either Messages.appError (\bms -> self.setState _ { breadcrumb = bms }) >>> liftEffect
