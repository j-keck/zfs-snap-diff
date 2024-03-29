module ZSD.BrowseSnapshots.CloneSnapshot where

import Prelude

import Data.Array ((..))
import Data.Array as A
import Data.Either (Either(..), either)
import Data.Enum (toEnumWithDefaults)
import Data.Foldable (foldMap)
import Data.Maybe (Maybe(..), fromMaybe, isJust)
import Data.Monoid (guard)
import Data.String as S
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import React.Basic (JSX)
import React.Basic.Classic (Component, Self, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture, capture_, key, targetValue)
import React.Basic.Events (handler)
import ZSD.Components.Confirm as Confirm
import ZSD.Components.Spinner as Spinner
import ZSD.Fragments.FormCommandFlag (flag)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Dataset as Dataset
import ZSD.Model.Snapshot (Snapshot)
import ZSD.Views.Messages as Messages

type Props
  = { dataset :: Dataset
    , snap :: Snapshot
    , onOk :: Effect Unit
    , onCancel :: Effect Unit
    }

type Flags
  = Array String

type State
  = { base :: String
    , flags :: Flags
    , fsName :: Maybe String
    , error :: Maybe String
    }

data Action
  = UpdateFsName String
  | CloneSnapshot

update :: Self Props State -> Action -> Effect Unit
update self = case _ of
  UpdateFsName name ->
    either (\error -> self.setState _ { error = Just error, fsName = Nothing })
      (\fsName -> self.setState _ { fsName = Just fsName, error = Nothing })
      $ validateName name
  CloneSnapshot -> flip foldMap self.state.fsName \name ->
       Spinner.display *> launchAff_ do
          let fsName = self.state.base <> "/" <> name
          res <- Dataset.cloneSnapshot self.props.dataset self.props.snap self.state.flags fsName
          liftEffect $ self.props.onOk *> Spinner.remove *> either Messages.appError Messages.info res

cloneSnapshot :: Props -> JSX
cloneSnapshot = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "CloneSnapshot"

  initialState = { base: "", fsName: Nothing, flags: [], error: Nothing }

  didMount self = self.setState _ { base = S.takeWhile ((/=) $ S.codePointFromChar '/') self.props.dataset.name }

  render self =
    Confirm.confirm
      { header: R.text "Clone snapshot"
      , body:
        R.form
          { onSubmit: capture_ $ pure unit
          , children:
            [ R.p { className: "font-weight-bold", children: [ R.text self.props.snap.fullName ] }
            , flag "-p"
                "Creates all the non-existing parent datasets.  Datasets created in this manner are automatically mounted according to the mountpoint property inherited from their parent.  If the target filesystem or volume already exists, the operation completes successfully."
                (addOrRemoveFlag self "-p")
            , R.div
                { className: "input-group mt-4"
                , children:
                  [ R.div
                      { className: "input-group-prepend"
                      , children: [ R.div { className: "input-group-text", children: [ R.text $ self.state.base <> "/" ] } ]
                      }
                  , R.input
                      { className: "form-control" <> guard (isJust self.state.error) " is-invalid"
                      , id: "fs-name"
                      , placeholder: "Filesystem name"
                      , autoFocus: true
                      , onChange: capture targetValue (fromMaybe "" >>> UpdateFsName >>> update self)
                      , onKeyDown: handler key $
                        case _ of
                          Just "Enter" -> update self CloneSnapshot
                          Just "Escape" -> self.props.onCancel
                          _ -> pure unit
                      }
                  , flip foldMap self.state.error
                      ( \error ->
                          R.div
                            { className: "invalid-feedback"
                            , children: [ R.text error ]
                            }
                      )
                  ]
                }
            ]
          }
      , onOk: update self CloneSnapshot
      , onCancel: self.props.onCancel
      }

  addOrRemoveFlag self flag true = self.setState _ { flags = A.snoc self.state.flags flag }

  addOrRemoveFlag self flag false = self.setState _ { flags = A.filter ((/=) flag) self.state.flags }

validateName :: String -> Either String String
validateName name =
  -- https://wiki.openindiana.org/oi/ZFS+naming+conventions
  let
    validStrs =
      (toEnumWithDefaults bottom top >>> S.singleton <$> 48 .. 57 <> 65 .. 90 <> 97 .. 122)
        <> [ "/", "_", "-", ":", "." ]

    invalidStrs = A.filter (flip A.elem validStrs >>> not) $ (S.singleton <$> S.toCodePointArray name)
  in
    if S.null name then
      Left $ "Name can't be empty"
    else if A.null invalidStrs then
      Right name
    else
      Left $ "Invalid character found: " <> (S.joinWith ", " invalidStrs)
