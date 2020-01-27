module ZSD.Fragments.FileActions.ViewText where

import Effect (Effect)
import Prelude (Unit, unit)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R

type Props = { content :: String }

foreign import highlightCode :: Effect Unit

viewText :: Props -> JSX
viewText = make component { initialState, render, didMount, didUpdate }

  where
    
    component :: Component Props
    component = createComponent "ViewText"

    initialState = unit
    
    didMount _ = highlightCode

    didUpdate _ _ = highlightCode

    render self = 
      R.pre_ [ R.code_ [ R.text self.props.content ] ]
