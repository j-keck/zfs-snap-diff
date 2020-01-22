module ZSD.Components.FileAction.ViewDiff where

import Prelude

import Data.Either (fromRight)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..))
import Data.Monoid (guard)
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, fragment, make)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import ZSD.Components.ActionButton (actionButton)
import ZSD.Model.Diff (Diff)
import ZSD.Model.Diff as Diff
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..))
import ZSD.Ops (zipWithIndex)


type Self = React.Self Props State

type Props =
  { file :: FSEntry
  , version :: FileVersion
  }

type State =
  { view :: View
  , diff :: Maybe Diff
  }

data View =
    InlineDiff
  | SideBySideDiff

instance showView :: Show View where
  show InlineDiff = "Inline"
  show SideBySideDiff = "Side by side"
derive instance eqView :: Eq View

data Command =
    Diff
  | Revert Int

update :: Self -> Command -> Effect Unit
update self = case _ of
  Diff -> case self.props.version of
    ActualVersion _ -> self.setState _ { diff = Nothing }
    BackupVersion _ -> launchAff_ $ do
      res <- unsafePartial $ fromRight <$> Diff.fetch self.props.file self.props.version
      liftEffect $ self.setState _ { diff = Just res }

  Revert idx -> launchAff_ $ do
    _ <- Diff.revert self.props.file self.props.version idx
    liftEffect $ update self Diff


viewDiff :: Props -> JSX
viewDiff = make component { initialState, render, didMount, didUpdate }
  where

    component :: Component Props
    component = createComponent "ViewDiff"

    initialState = { view: InlineDiff, diff: Nothing }

    didMount self = update self Diff

    didUpdate self {prevState, prevProps } =
      guard (self.props.version /= prevProps.version) $ update self Diff

    render self =
      fragment
      [ R.ul
        { className: "nav nav-tabs"
        , children:
          [ mkTabEntry InlineDiff     self
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
      , children: [ R.a { className: "nav-link" <> guard (self.state.view == id) " active"
                        , href: "#"
                        , onClick: capture_ $ self.setState _ { view = id }
                        , children: [ R.text $ show id ]
                        }
                  ]
      }

    inlineDiff self =
      flip foldMap self.state.diff \diff -> fragment $
        flip map (zipWithIndex diff.inlineDiffHTMLFragment) \(Tuple idx html) ->
          R.div
          { className: "m-1"
          , children:
            [ actionButton
              { text: "Revert"
              , textConfirm: "Revert this change"
              , icon: ""
              , action: update self (Revert idx)
              }
            , R.pre_ [ R.code { dangerouslySetInnerHTML: { __html: html } } ]
            ]
          }
        


    sideBySideDiff self =
      flip foldMap self.state.diff \diff -> fragment $
        flip map (zipWithIndex diff.sideBySideDiffHTMLFragment) \(Tuple idx html) ->
          R.div
          { className: "m-1"
          , children:
            [ actionButton
              { text: "Revert"
              , textConfirm: "Revert this change"
              , icon: ""
              , action: update self (Revert idx)
              }
            , R.table
              { className: "table table-borderless table-sm"
              , dangerouslySetInnerHTML: { __html: html }
              }
            ]
          }


