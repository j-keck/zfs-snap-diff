module ZSD.Components.DirBrowser where

import Prelude
import Data.Array (snoc)
import Data.Array as A
import Data.Either (fromRight)
import Data.Monoid (guard)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
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
  { dataset         :: Dataset
  , onFileSelected  :: FSEntry -> Effect Unit
  , onDirSelected   :: FSEntry -> Effect Unit
  }

type State =
  { breadcrumb     :: Array FSEntry
  , dirListing     :: Array FSEntry
  , showBrowser    :: Boolean
  , showHidden     :: Boolean
  }

type Self = React.Self Props State

data Actions =
    StartAt FSEntry
  | ChangeDir FSEntry
  | PickFromBreadcrumb FSEntry
  | OnClick FSEntry
  | ReadDir FSEntry


dirBrowser :: Props -> JSX
dirBrowser = make component { initialState, render, didMount, didUpdate }

  where

    component :: Component Props
    component  = createComponent "DirBrowser"

    initialState = { breadcrumb: [], dirListing: [], showBrowser: true, showHidden: false }

    didMount self = send self (StartAt self.props.dataset.mountPoint)

    didUpdate self {prevProps} = do
      guard (self.props.dataset /= prevProps.dataset) $
        send self (StartAt self.props.dataset.mountPoint)

    send self = case _ of
      StartAt target -> do
        send self $ ReadDir target
        self.setState _ { breadcrumb = [target] }

      ChangeDir target -> do
        send self $ ReadDir target
        self.setState \s -> s { breadcrumb = s.breadcrumb `snoc` target }

      PickFromBreadcrumb target -> do
        send self $ ReadDir target
        let breadcrumb = A.takeWhile (_ /= target) self.state.breadcrumb
        self.setState _ { breadcrumb = breadcrumb `snoc` target, showBrowser = true }

      ReadDir fh -> launchAff_ $ do
        dirListing <- DirListing.fetch fh
        -- FIXME: handle errors
        liftEffect $ self.setState _ { dirListing = unsafePartial $ fromRight dirListing }


      OnClick fsh -> do
        case fsh.kind of
          "DIR" -> do
            -- FIXME: spinning modal
            -- self.setState _ { dirListing = [] }
            send self $ ChangeDir fsh
            self.props.onDirSelected fsh
          "FILE" -> do
            self.setState _ { showBrowser = false }
            self.props.onFileSelected fsh
          _ -> pure unit





    render self =
      R.div
      { className: "mt-3"
      , children:
        [ breadcrumb
        , guard self.state.showBrowser browser
        ]
      }


      where breadcrumb =
              R.nav_
              [ R.ol
                 { className: "breadcrumb"
                 , children: map (\h -> R.li
                                      { className: "breadcrumb-item"
                                      , children:
                                        [ R.a
                                          { onClick: capture_ $ send self (PickFromBreadcrumb h)
                                          , href: "#"
                                          , children: [ R.text h.name ]
                                          }
                                        ]
                                      }
                                  ) self.state.breadcrumb
                 }
              ]


            browser =
              table
              { header: ["Name", "Size", "Modify time"]
              , rows: DirListing.filter self.state self.state.dirListing
              , mkRow: \f -> [ R.img { className: "m-2", src: icon f } <> R.text f.name
                            , R.text $ Formatter.filesize f.size
                            , R.text $ Formatter.dateTime f.modTime
                            ]
              , onRowSelected: send self <<< OnClick
              }

            icon e
              | FSEntry.isFile e = "icons/file.svg"
              | FSEntry.isDir e = "icons/file-directory.svg"
              | FSEntry.isLink e = "icons/file-symlink-file.svg"
              | otherwise = "icons/database.svg"


