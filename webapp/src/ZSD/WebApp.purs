-- | WebApp entry point
-- |
module ZSD.WebApp where

import Data.Array as A
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Prelude (map, ($), (/=))
import React.Basic (Component, JSX, createComponent, fragment, make)
import React.Basic.DOM as R
import ZSD.Components.Messages as Messages
import ZSD.Components.Navbar (navbar)
import ZSD.Model.Config (Config)
import ZSD.View.BrowseSnapshots (browseSnapshots)
import ZSD.Views.BrowseFilesystem (browseFilesystem)


type Props = { config :: Config }

type Title = String
type View = JSX

type State =
  { views       :: Array (Tuple Title View)
  , activeTitle :: Title
  }


webApp :: Props -> JSX
webApp props = make component { initialState, render } props

  where

    component :: Component Props
    component = createComponent "WebApp"


    initialState =
      let views = [ Tuple "Browse filesystem" $ browseFilesystem { config: props.config }
                  , Tuple "Browse snapshots" $ browseSnapshots { config: props.config }
                  , Tuple "Messages" $ Messages.messages
                  ]
                  
       in { views, activeTitle: "" } 


    render self = fragment $ 
      A.concat
      [ [ navbar
          { views: self.state.views
          , onViewSelected: \title -> self.setState _ { activeTitle = title }
          }
        , Messages.toasts
        ]
        , map (embedView self) self.state.views
      ]

    embedView self (Tuple title view) =
      R.div
      { className: guard (title /= self.state.activeTitle) "d-none"
      , children: [ view ]
      }
