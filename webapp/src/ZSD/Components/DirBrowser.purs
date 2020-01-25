module ZSD.Components.DirBrowser where

import Prelude

import Data.Array (snoc)
import Data.Array as A
import Data.Either (fromRight)
import Data.Maybe (Maybe(..), maybe)
import Data.Monoid (guard)
import Data.Tuple (snd)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Component.Table (table)
import ZSD.Formatter as Formatter
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.DirListing as DirListing
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FSEntry as FSEntry


type Props =
  { path            :: FSEntry
  , onFileSelected  :: FSEntry -> Effect Unit
  , onDirSelected   :: FSEntry -> Effect Unit
  }

type State =
  { breadcrumb     :: Array FSEntry
  , dirListing     :: Array FSEntry
  , selectedFile   :: Maybe FSEntry
  , showBrowser    :: Boolean
  , showHidden     :: Boolean
  }

type Self = React.Self Props State

data Command =
    StartAt FSEntry
  | ChangeDir FSEntry
  | PickFromBreadcrumb FSEntry
  | OnClick FSEntry
  | ReadDir FSEntry

update :: Self -> Command -> Effect Unit
update self = case _ of
  StartAt target -> do
    update self $ ReadDir target
    self.setState _ { breadcrumb = [target] }

  ChangeDir target -> do
    update self $ ReadDir target
    self.setState \s -> s { breadcrumb = s.breadcrumb `snoc` target, selectedFile = Nothing }

  PickFromBreadcrumb target -> do
    update self $ ReadDir target
    self.props.onDirSelected target
    let breadcrumb = A.takeWhile (_ /= target) self.state.breadcrumb
    self.setState _ { breadcrumb = breadcrumb `snoc` target
                    , showBrowser = true, selectedFile = Nothing }

  ReadDir fh -> launchAff_ $ do
    dirListing <- DirListing.fetch fh
    -- FIXME: handle errors
    liftEffect $ self.setState _ { dirListing = unsafePartial $ fromRight dirListing }


  OnClick fsh -> do
    case fsh.kind of
      "DIR" -> do
        -- FIXME: spinning modal
        -- self.setState _ { dirListing = [] }
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
                   , showBrowser: true, showHidden: false }

    didMount self = update self (StartAt self.props.path)

    didUpdate self {prevProps} = do
      guard (self.props.path /= prevProps.path) $
        update self (StartAt self.props.path)


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
          , children: map (\h -> R.li
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
          }
       ]




