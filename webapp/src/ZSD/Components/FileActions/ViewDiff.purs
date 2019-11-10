module ZSD.Components.FileAction.ViewDiff where

import Prelude

import React.Basic (Component, JSX, createComponent, fragment, make)
import React.Basic.DOM as R
import ZSD.Model.Diff (Diff)


type Props = { diff :: Diff }


viewDiff :: Props -> JSX
viewDiff = make component { initialState, render }
  where

    component :: Component Props
    component = createComponent "ViewDiff"

    initialState = unit

    render self = fragment $ flip map self.props.diff.sideBySide \html ->
      R.table
      { className: "table table-borderless table-sm"
      , dangerouslySetInnerHTML: { __html: html }
      }

