module ZSD.Fragments.CreateSnapshotModal where

import Prelude
import Data.Array as A
import Data.Array ((..))
import Data.DateTime (DateTime)
import Data.Either (Either(..), either)
import Data.Foldable (fold)
import Data.Enum (toEnumWithDefaults)
import Data.Formatter.DateTime (FormatterCommand(..), format)
import Data.List as List
import Data.Maybe (fromMaybe)
import Data.Monoid (guard)
import Data.String as S
import Data.Traversable (traverse)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Now (nowDateTime)
import React.Basic (Component, JSX, Self, createComponent, fragment, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture, capture_, targetValue)
import ZSD.Model.Dataset (Dataset)
import ZSD.Model.Dataset as Dataset
import ZSD.Views.Messages as Messages

type Props
  = { dataset :: Dataset
    , defaultTemplate :: String
    , onRequestClose :: Effect Unit
    }

type State
  = { snapshotTemplate :: String
    , snapshotName :: String
    , error :: String
    , showHelp :: Boolean
    }

data Action
  = CreateSnapshot
  | ConvertTemplate String

update :: Self Props State -> Action -> Effect Unit
update self = case _ of
  ConvertTemplate s -> do
    self.setState _ { snapshotTemplate = s }
    ts <- nowDateTime
    either (\error -> self.setState _ { error = error })
      (\name -> self.setState _ { snapshotName = name, error = "" })
      $ (convert ts s >>= validateName)
  CreateSnapshot ->
    guard (S.null self.state.error)
      $ launchAff_
      $ Dataset.createSnapshot self.props.dataset self.state.snapshotName
      >>= (\res -> liftEffect $ either Messages.appError Messages.info res *> self.props.onRequestClose)

createSnapshotModal :: Props -> JSX
createSnapshotModal = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "SnapshotNameModal"

  initialState = { snapshotTemplate: "", snapshotName: "", error: "", showHelp: false }

  didMount self =
    self.setState _ { snapshotTemplate = self.props.defaultTemplate }
      *> update self (ConvertTemplate self.props.defaultTemplate)

  render self =
    fragment
      [ R.div
          { className: "modal modal-show"
          , children:
            [ div "modal-dialog modal-dialog-centered"
                $ div "modal-content"
                $ fragment
                    [ div "modal-header" $ R.text "Create ZFS Snapshot"
                    , div "modal-body m-1" $ body self
                    , div "modal-footer"
                        $ fragment
                            [ R.button
                                { className: "btn btn-secondary"
                                , onClick: capture_ self.props.onRequestClose
                                , children: [ R.text "Cancel" ]
                                }
                            , R.button
                                { className:
                                  "btn btn-primary"
                                    <> guard (not $ S.null self.state.error) " disabled"
                                , onClick: capture_ $ update self CreateSnapshot
                                , children: [ R.text "Create" ]
                                }
                            ]
                    ]
            ]
          }
      , R.div { className: "modal-backdrop fade show" }
      ]

  div className child = R.div { className, children: [ child ] }

  body self =
    R.form
      { children:
        [ div "form-group row"
            $ R.label
                { htmlFor: "template"
                , children: [ R.text "Snapshot name template" ]
                }
            <> R.input
                { className:
                  "form-control"
                    <> guard (not $ S.null self.state.error) " is-invalid"
                , id: "template"
                , placeholder: "Snapshot name template"
                , onChange: capture targetValue (fromMaybe "" >>> ConvertTemplate >>> update self)
                , value: self.state.snapshotTemplate
                }
            <> div "invalid-feedback" (R.text self.state.error)
            <> R.small
                { className: "form-text pointer"
                , onClick: capture_ $ self.setState \s -> s { showHelp = not s.showHelp }
                , children: [ R.text "Show supported format sequences." ]
                }
            <> R.pre
                { className: "slide" <> guard (not self.state.showHelp) " d-none"
                , children:
                  map ((<>) "\n" >>> R.text)
                    [ "Format sequences are alike the `date` command"
                    , "  %d: day of month (e.g., 01)"
                    , "  %m: month (01..12)"
                    , "  %y: last two digits of year (00..99)"
                    , "  %Y: year"
                    , "  %F: full date; like %Y-%m-%d"
                    , "  %H: hour (00..23)"
                    , "  %I: hour (01..12)"
                    , "  %M: minute (00..59)"
                    , "  %S: second (00..60)"
                    , "  %s: seconds since 1970-01-01 00:00:00 UTC"
                    ]
                }
        , div "form-group row"
            $ R.label
                { htmlFor: "name"
                , children: [ R.text "Snapshot name" ]
                }
            <> R.input
                { className: "form-control"
                , id: "name"
                , readOnly: true
                , value: self.state.snapshotName
                }
        ]
      }

validateName :: String -> Either String String
validateName name =
  -- https://wiki.openindiana.org/oi/ZFS+naming+conventions
  let
    validStrs =
      (toEnumWithDefaults bottom top >>> S.singleton <$> 48 .. 57 <> 65 .. 90 <> 97 .. 122)
        <> [ "_", "-", ":", "." ]

    invalidStrs = A.filter (flip A.elem validStrs >>> not) $ (S.singleton <$> S.toCodePointArray name)
  in
    if A.null invalidStrs then
      Right name
    else
      Left $ "Invalid character found: " <> (S.joinWith ", " invalidStrs)

convert :: DateTime -> String -> Either String String
convert dt = map fold <<< traverse fmt <<< parse
  where
  fmt = case _ of
    Plain s -> Right s
    Fmt cmds -> Right $ format (List.fromFoldable cmds) dt
    Invalid s -> Left s

data Frag
  = Plain String
  | Fmt (Array FormatterCommand)
  | Invalid String

parse :: String -> Array Frag
parse s
  | S.null s = []
  | S.take 1 s == "%" =
    A.cons (either Invalid Fmt $ decode $ S.take 2 s)
      (parse $ S.drop 2 s)
  | otherwise =
    let
      plain = S.takeWhile (_ /= S.codePointFromChar '%') s
    in
      A.cons (Plain plain) (parse $ S.drop (S.length plain) s)

decode :: String -> Either String (Array FormatterCommand)
decode = case _ of
  "%d" -> Right [ DayOfMonthTwoDigits ]
  "%m" -> Right [ MonthTwoDigits ]
  "%y" -> Right [ YearTwoDigits ]
  "%Y" -> Right [ YearFull ]
  "%F" ->
    Right
      [ YearFull
      , Placeholder "-"
      , MonthTwoDigits
      , Placeholder "-"
      , DayOfMonthTwoDigits
      ]
  "%H" -> Right [ Hours24 ]
  "%I" -> Right [ Hours12 ]
  "%M" -> Right [ MinutesTwoDigits ]
  "%S" -> Right [ SecondsTwoDigits ]
  "%s" -> Right [ UnixTimestamp ]
  s -> Left $ "Invalid format: " <> s
