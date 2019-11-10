module ZSD.Components.FileActions.ViewText where

import React.Basic (Component, JSX, createComponent, makeStateless)
import React.Basic.DOM as R

type Props = { content :: String }

viewText :: Props -> JSX
viewText = makeStateless component \props ->

  R.pre_ [ R.code_ [ R.text props.content ] ]

  where
    component :: Component Props
    component = createComponent "ViewText"
