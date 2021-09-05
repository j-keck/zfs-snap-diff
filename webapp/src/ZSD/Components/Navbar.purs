module ZSD.Components.Navbar where

import Prelude
import Data.Array as A
import Data.Foldable (foldMap)
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)

type Title
  = String

type View
  = JSX

type Props
  = { views :: Array (Tuple Title View)
    , onViewSelected :: View -> Effect Unit
    }

type State
  = { activeViewTitle :: Title }

navbar :: Props -> JSX
navbar = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "Navbar"

  initialState = { activeViewTitle: "" }

  didMount self =
    foldMap
      ( \(Tuple title view) ->
          self.setState _ { activeViewTitle = title }
            *> self.props.onViewSelected view
      )
      $ A.head self.props.views

  render self =
    R.nav
      { className: "navbar navbar-expand navbar-dark bg-primary"
      , children:
        [ navbarBrand
        , navbarItems self
        ]
      }

  navbarBrand =
    R.a
      { className: "navbar-brand"
      , target: "_blank"
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
            , onClick: capture_ $ self.setStateThen _ { activeViewTitle = title } (self.props.onViewSelected view)
            , children: [ R.text title ]
            }
        ]
      }
