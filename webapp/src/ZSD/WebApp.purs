-- | WebApp entry point
-- |
module ZSD.WebApp where

import Data.Array.NonEmpty (NonEmptyArray)
import Data.Array.NonEmpty as ANE
import Data.Tuple (Tuple(..), snd)
import Prelude (($))
import React.Basic (Component, JSX, createComponent, fragment, make)
import ZSD.Components.Navbar (navbar)
import ZSD.Model.Config (Config)
import ZSD.View.BrowseSnapshots (browseSnapshots)
import ZSD.Views.BrowseFilesystem (browseFilesystem)


type Props = { config :: Config }

type Title = String
type View = JSX

type State =
  { views :: NonEmptyArray (Tuple Title View)
  , activeView :: View
  }


webApp :: Props -> JSX
webApp props = make component { initialState, render } props

  where

    component :: Component Props
    component = createComponent "WebApp"


    initialState =
      let views = ANE.cons'
                  (Tuple "Browse filesystem"  $ browseFilesystem { config: props.config })
                  [ Tuple "Browse snapshots" $ browseSnapshots { config: props.config }]
       in { views, activeView: snd $ ANE.head views }


    render self =
      fragment
      [ navbar
        { views: self.state.views
        , onViewSelected: \view -> self.setState _ { activeView = view }
        }
      , self.state.activeView
      ]
