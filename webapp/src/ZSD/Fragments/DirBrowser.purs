module ZSD.Fragments.DirBrowser where

import Prelude

import Data.Array (snoc)
import Data.Array as A
import Data.Either (Either(..), either)
import Data.Maybe (Maybe(..), fromMaybe, maybe)
import Data.Monoid (guard)
import Data.String as S
import Data.Traversable as T
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Foreign.Object as O
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Component.Table (table)
import ZSD.Components.Messages as Messages
import ZSD.Components.Scroll as Scroll
import ZSD.Formatter as Formatter
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.DirListing as DirListing
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FSEntry as FSEntry
import ZSD.Ops (unsafeFromJust, (</>))
import ZSD.Views.BookmarkManager as BM


type Props =
  { ds              :: Dataset
  , root            :: FSEntry
  , onFileSelected  :: FSEntry -> Effect Unit
  , onDirSelected   :: FSEntry -> Effect Unit
  }

type State =
  { breadcrumb     :: Array FSEntry
  , dirListing     :: Array FSEntry
  , selectedFile   :: Maybe FSEntry
  , showBrowser    :: Boolean
  , showHidden     :: Boolean
  , bookmarks      :: Array String
  }

type Self = React.Self Props State

data Command =
    StartAt FSEntry
  | ChangeDir FSEntry
  | PickFromBreadcrumb FSEntry
  | PickFromBookmark String
  | OnClick FSEntry
  | ReadDir FSEntry

update :: Self -> Command -> Effect Unit
update self = case _ of
  
  StartAt target -> do
    update self $ ReadDir target
    bms <- BM.getBookmarks self.props.ds
    self.setState _ { breadcrumb = [target], selectedFile = Nothing, bookmarks = bms }


  ChangeDir target -> do
    update self $ ReadDir target
    self.setState \s -> s { breadcrumb = s.breadcrumb `snoc` target, selectedFile = Nothing }

  
  PickFromBreadcrumb target -> do
    let breadcrumb = A.takeWhile (_ /= target) self.state.breadcrumb
    self.setState _ { breadcrumb = breadcrumb `snoc` target
                    , showBrowser = true, selectedFile = Nothing }
    update self $ ReadDir target


  PickFromBookmark path -> do
    launchAff_ $ do
      let pathElements = T.scanl (</>) "" (S.split (S.Pattern "/") $ self.props.root.path <>  path)
      T.sequence <$> T.traverse FSEntry.stat pathElements >>= (case _ of
        Left err -> Messages.appError err
        Right ps -> do
          self.setState _ { breadcrumb = A.dropWhile ((/=) self.props.root) ps
                          , selectedFile = Nothing
                          }
          update self $ ReadDir (unsafeFromJust $ A.last ps)
        ) >>> liftEffect

        
  ReadDir fh -> launchAff_ do
    res <- DirListing.fetch fh
    liftEffect $ either Messages.appError (\ls -> self.setState _ { dirListing = ls }) res
   

  OnClick fsh -> Scroll.scrollToTop *> do
    case fsh.kind of
      "DIR" -> do
        -- FIXME: spinning modal?
        update self $ ChangeDir fsh
        self.props.onDirSelected fsh
      "FILE" -> do
        self.setState _ { showBrowser = false, selectedFile = Just fsh }
        self.props.onFileSelected fsh
      _ -> pure unit



dirBrowser :: Props -> JSX
dirBrowser = make component { initialState, render, didMount, didUpdate }

  where

    component :: Component Props
    component  = createComponent "DirBrowser"

    initialState = { breadcrumb: [], dirListing: [], selectedFile: Nothing
                   , showBrowser: true, showHidden: false, bookmarks: []}

    didMount self = update self (StartAt self.props.root)

    didUpdate self {prevProps} = do
      guard (self.props.root /= prevProps.root) $ 
        update self (StartAt self.props.root)


    render self =
      R.div
      { className: "mt-3"
      , children:
        [ breadcrumb self
        , guard self.state.showBrowser $
             table
              { header: ["Name", "Size", "Modify time"]
              , rows: DirListing.filter self.state self.state.dirListing
              , mkRow: \f -> [ R.span { className: icon f } <> R.text f.name
                            , R.text $ Formatter.filesize f.size
                            , R.text $ Formatter.dateTime f.modTime
                            ]
              , onRowSelected: update self <<< OnClick
              }
        ]
      }


    icon e
      | FSEntry.isFile e = "fas fa-file p-1"
      | FSEntry.isDir e = "fas fa-folder p-1"
      | FSEntry.isLink e = "fas fa-link p-1"
      | otherwise = "fas fa-hdd p-1"


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
                             , children: [ R.text h.name ]
                             }
                           ]
                         }
                  ) self.state.breadcrumb
              `A.snoc` (maybe mempty (\f -> R.li
                                            { className: "breadcrumb-item"
                                            , children: [ R.text f.name ]
                                            })
                        $ self.state.selectedFile)
            , [ R.div
                { className: "ml-auto"
                , children:
                  [ R.span
                    { className: "dropdown"
                    , children:
                      [ R.button
                        { className: "btn btn-secondary py-1"
                        , onClick: capture_ $ (if isCurrentDirBookmarked self
                                               then BM.removeBookmark self.props.ds (currentDir self)
                                               else BM.addBookmark self.props.ds (currentDir self))
                                      >>= \bms -> self.setState _ { bookmarks = bms }
                        , children: [ let fa = if isCurrentDirBookmarked self then "fas" else "far" in
                                       R.span { className: fa <> " fa-bookmark px-2" }
                                    ]                          
                        }
                      , R.button
                        { className: "btn btn-secondary py-1 dropdown-toggle dropdown-toggle-split"
                        , _data: O.fromHomogeneous {toggle: "dropdown" }
                        }
                      , R.div
                        { className: "dropdown-menu"
                        , children: map (\b -> R.a { className: "dropdown-item"
                                                   , href: "#"
                                                   , onClick: capture_ $ update self $ PickFromBookmark b
                                                   , children: [ R.text b ]
                                                   }) self.state.bookmarks
                        }
                      ]
                    }
                  ]
                }
              ]
            ]
          }
       ]


    currentDir self = dropRoot $ (fromMaybe self.props.root $ A.last self.state.breadcrumb).path
      where dropRoot p = fromMaybe p $ S.stripPrefix (S.Pattern self.props.root.path) p
            
    isCurrentDirBookmarked self = A.elem (currentDir self) self.state.bookmarks
