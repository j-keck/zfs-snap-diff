module ZSD.Fragments.FileAction.ViewDiff where

import Prelude
import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (JSX)
import React.Basic.Classic (Component, createComponent, fragment, make)
import React.Basic.Classic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Components.ActionButton (actionButton)
import ZSD.Views.Messages as Messages
import ZSD.Model.Diff (Diff)
import ZSD.Model.Diff as Diff
import ZSD.Model.FH (FH)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Utils.Ops (zipWithIndex)

type Self
  = React.Self Props State

type Props
  = { file :: FH
    , version :: FileVersion
    }

type State
  = { view :: View
    , diff :: Maybe Diff
    }

data View
  = InlineDiff
  | SideBySideDiff

instance showView :: Show View where
  show InlineDiff = "Inline"
  show SideBySideDiff = "Side by side"

derive instance eqView :: Eq View

data Command
  = Diff
  | Revert Int

update :: Self -> Command -> Effect Unit
update self = case _ of
  Diff -> case self.props.version of
    CurrentVersion _ -> self.setState _ { diff = Nothing }
    BackupVersion _ ->
      launchAff_
        $ do
            res <- Diff.fetch self.props.version
            liftEffect $ either Messages.appError (\diff -> self.setState _ { diff = Just diff }) res
  Revert idx ->
    launchAff_
      $ do
          res <- Diff.revert self.props.version idx
          liftEffect
            $ do
                either Messages.appError Messages.info res
                update self Diff

viewDiff :: Props -> JSX
viewDiff = make component { initialState, render, didMount, didUpdate }
  where
  component :: Component Props
  component = createComponent "ViewDiff"

  initialState = { view: InlineDiff, diff: Nothing }

  didMount self = update self Diff

  didUpdate self { prevState: _, prevProps } = guard (self.props.version /= prevProps.version) $ update self Diff

  render self = case self.props.version of
    CurrentVersion _ ->
      R.p
        { className: "text-center font-weight-bold"
        , children: [ R.text "Current version selected - select an older version to view a diff." ]
        }
    BackupVersion _ ->
      fragment
        [ R.ul
            { className: "nav nav-tabs"
            , children:
              [ mkTabEntry InlineDiff self
              , mkTabEntry SideBySideDiff self
              ]
            }
        , case self.state.view of
            InlineDiff -> inlineDiff self
            SideBySideDiff -> sideBySideDiff self
        ]

  mkTabEntry id self =
    R.li
      { className: "nav-item"
      , children:
        [ R.a
            { className: "nav-link" <> guard (self.state.view == id) " active"
            , href: "#"
            , onClick: capture_ $ self.setState _ { view = id }
            , children: [ R.text $ show id ]
            }
        ]
      }

  inlineDiff self =
    flip foldMap self.state.diff \diff ->
      fragment
        $ flip map (zipWithIndex diff.inlineDiffHTMLFragment) \(Tuple idx html) ->
            R.div
              { className: "m-1"
              , children:
                [ actionButton
                    { text: "Revert"
                    , title: "Revert this change"
                    , textConfirm: "Revert this change"
                    , icon: ""
                    , action: update self (Revert idx)
                    , enabled: true
                    }
                , R.pre_ [ R.code { dangerouslySetInnerHTML: { __html: html } } ]
                ]
              }

  sideBySideDiff self =
    flip foldMap self.state.diff \diff ->
      fragment
        $ flip map (zipWithIndex diff.sideBySideDiffHTMLFragment) \(Tuple idx html) ->
            R.div
              { className: "m-1"
              , children:
                [ actionButton
                    { text: "Revert"
                    , title: "Revert this change"
                    , textConfirm: "Revert this change"
                    , icon: ""
                    , action: update self (Revert idx)
                    , enabled: true
                    }
                , R.table
                    { className: "table table-borderless table-sm"
                    , dangerouslySetInnerHTML: { __html: html }
                    }
                ]
              }
