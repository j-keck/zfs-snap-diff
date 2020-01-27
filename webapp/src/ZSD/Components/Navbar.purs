module ZSD.Components.Navbar where

import Data.Array as A
import Data.Monoid (guard)
import Data.Tuple (Tuple(..), fst)
import Effect (Effect)
import Prelude (Unit, map, ($), (<>), (==))
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Ops (unsafeFromJust)


type Title = String
type View = JSX

type Props =
  { views :: Array (Tuple Title View)
  , onViewSelected :: Title -> Effect Unit
  }

type State = { activeViewTitle :: Title }

navbar :: Props -> JSX
navbar = make component { initialState, didMount, render } 

  where

    component :: Component Props
    component = createComponent "Navbar"

    initialState = { activeViewTitle: "" }

    didMount self =
      let title = fst $ unsafeFromJust $ A.head self.props.views in
      self.setStateThen _ { activeViewTitle = title } $ self.props.onViewSelected title


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
          , children: map (mkNavItem self) $ self.props.views
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
          , onClick: capture_ $ self.setStateThen _ { activeViewTitle = title } (self.props.onViewSelected title)
          , children: [ R.text title ]
          }
        ]
      }
