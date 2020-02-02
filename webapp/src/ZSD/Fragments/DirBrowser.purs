module ZSD.Fragments.DirBrowser where

import Prelude

import Data.Array as A
import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.Monoid (guard)
import Data.Newtype (unwrap)
import Data.String as S
import Data.Traversable as T
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Unsafe (unsafePerformEffect)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import Unsafe.Coerce (unsafeCoerce)
import ZSD.Components.DropDownButton (dropDownButton)
import ZSD.Components.Scroll as Scroll
import ZSD.Components.Spinner as Spinner
import ZSD.Components.Table (table)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.FH (FH(..), From(..), To(..), switchMountPoint)
import ZSD.Model.FH as FH
import ZSD.Model.Kind (Kind(..), icon)
import ZSD.Model.MountPoint (MountPoint)
import ZSD.Utils.BookmarkManager as BM
import ZSD.Utils.Formatter as Formatter
import ZSD.Utils.Ops (tupleM, (</>))
import ZSD.Views.Messages as Messages


type Props =
  { ds              :: Dataset
  , altRoot         :: Maybe MountPoint
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
                    (\alt -> switchMountPoint (From self.props.ds.mountPoint) (To alt) fh)
                    self.props.altRoot
    in    update self (ReadDir fh')
       *> self.setState _ { selectedFile = Nothing, showBrowser = true }
       *> self.props.onDirSelected fh
       *> rebuildBreadcrumb self fh'


  ReadDir fh ->
       Spinner.display
    *> self.setState _ { currentDir = fh }
    *> launchAff_ (    FH.ls fh
                   >>= either Messages.appError (\ls -> self.setState _ { dirListing = ls, currentDir = fh })
                   >>> liftEffect)
    *> Spinner.remove



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


  SwitchRoot old new ->
    let fh = switchMountPoint old new self.state.currentDir
    in     Spinner.display
        *> update self (ReadDir fh)
        *> self.setState _ { currentDir = fh }
        *> rebuildBreadcrumb self fh
        *> Spinner.remove


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


    didMount self = update self (StartAt $ fromMaybe self.props.ds.mountPoint self.props.altRoot)


    didUpdate self {prevProps} = do
      guard (self.props.ds /= prevProps.ds) $
        update self (StartAt $ fromMaybe self.props.ds.mountPoint self.props.altRoot)

      foldMap (\(Tuple old new) ->
                guard (old /= new) $ update self (SwitchRoot (From old) (To new)))
              (tupleM prevProps.altRoot self.props.altRoot)


    render self =
      R.div
      { className: "mt-3"
      , children:
        [ breadcrumb self
        , guard self.state.showBrowser $
             table
              { header: ["Name", "Size", "Modify time"]
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
                                    (\mp -> switchMountPoint (From mp) (To self.props.ds.mountPoint) self.state.currentDir)
                                    self.props.altRoot
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
  let root = fromMaybe self.props.ds.mountPoint self.props.altRoot
      pathElements = unwrap >>> _.path >>> S.split (S.Pattern "/") >>> T.scanl (</>) "" $ fh
      p = A.dropWhile ((/=) (unwrap >>>_.path $ root)) pathElements
  FH.stat' p >>= either Messages.appError (\bms -> self.setState _ { breadcrumb = bms }) >>> liftEffect
