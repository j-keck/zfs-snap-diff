module ZSD.Components.Navbar where

import Data.Array.NonEmpty (NonEmptyArray)
import Data.Array.NonEmpty as ANE
import Data.Monoid (guard)
import Data.Tuple (Tuple(..), fst)
import Effect (Effect)
import Prelude (Unit, map, ($), (<>), (==))
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)


type Title = String
type View = JSX

type Props =
  { views :: NonEmptyArray (Tuple Title View)
  , onViewSelected :: View -> Effect Unit
  }

type State = { activeViewTitle :: Title }

navbar :: Props -> JSX
navbar props = make component { initialState, render } props

  where

    component :: Component Props
    component = createComponent "Navbar"

    initialState = { activeViewTitle: fst $ ANE.head props.views }

    render self =
      R.nav
      { className: "navbar navbar-expand-lg navbar-dark bg-primary"
      , children:
        [ navbarBrand
        , navbarItems self
        ]
      }


    navbarBrand =
      R.a
      { className: "navbar-brand"
      , href: "https://j-keck.github.com/zfs-snap-diff"
      , children: [ R.text "ZFS-Snap-Diff" ]
      }


    navbarItems self =
      R.div
      { className: "collapse navbar-collapse"
      , children:
        [ R.ul
          { className: "navbar-nav"
          , children: ANE.toUnfoldable $ map (mkNavItem self) props.views
          }
        ]
      }


    mkNavItem self (Tuple title view) =
      R.li
      { className: "nav-item" <> guard (title == self.state.activeViewTitle) " active"
      , children:
        [ R.a
          { className: "nav-link"
          , href: "#"
          , onClick: capture_ $ self.setStateThen _ { activeViewTitle = title } (self.props.onViewSelected view)
          , children: [ R.text title ]
          }
        ]
      }
