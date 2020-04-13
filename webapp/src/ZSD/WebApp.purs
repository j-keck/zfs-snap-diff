-- | WebApp entry point
-- |
-- | Contains the navigation bar and the side content
-- |
module ZSD.WebApp where

import Data.Maybe (Maybe(..))
import Data.Tuple (Tuple(..))
import Prelude (($))
import React.Basic (Component, JSX, createComponent, empty, fragment, make)
import ZSD.Components.Navbar (navbar)
import ZSD.Components.Spinner as Spinner
import ZSD.Model.Config (Config)
import ZSD.Model.Dataset (Dataset)
import ZSD.Views.BrowseFilesystem (browseFilesystem)
import ZSD.Views.BrowseSnapshots (browseSnapshots)
import ZSD.Views.Messages as Messages

type Props
  = { config :: Config }

type Title
  = String

type View
  = JSX

type State
  = { views :: Array (Tuple Title View)
    , activeView :: View
    , activeDataset :: Maybe Dataset
    }

webApp :: Props -> JSX
webApp props = make component { initialState, render } props
  where
  component :: Component Props
  component = createComponent "WebApp"

  initialState = { views: [], activeView: empty, activeDataset: Nothing }

  render self =
    fragment
      $ [ navbar
            { views:
              [ Tuple "Browse filesystem"
                  $ browseFilesystem
                      { config: props.config
                      , activeDataset: self.state.activeDataset
                      , onDatasetSelected:
                        \ds -> self.setState _ { activeDataset = Just ds }
                      }
              , Tuple "Browse snapshots"
                  $ browseSnapshots
                      { config: props.config
                      , activeDataset: self.state.activeDataset
                      , onDatasetSelected:
                        \ds -> self.setState _ { activeDataset = Just ds }
                      }
              , Tuple "Messages" $ Messages.messages
              ]
            , onViewSelected: \view -> self.setState _ { activeView = view }
            }
        , Messages.toasts
        , self.state.activeView
        , Spinner.spinner
        ]
