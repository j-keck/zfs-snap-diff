module ZSD.Fragments.SnapshotNameForm where

import Prelude

import Data.Array ((..))
import Data.Array as A
import Data.DateTime (DateTime)
import Data.Either (Either(..), either)
import Data.Enum (toEnumWithDefaults)
import Data.Foldable (fold, foldMap)
import Data.Formatter.DateTime (FormatterCommand(..), format)
import Data.List as List
import Data.Maybe (Maybe(..), fromMaybe, isJust)
import Data.Monoid (guard)
import Data.String as S
import Data.Traversable (traverse)
import Effect (Effect)
import Effect.Now (nowDateTime)
import React.Basic (JSX)
import React.Basic.Classic (Component, Self, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture, capture_, key, targetValue)
import React.Basic.Events (handler)
import ZSD.Model.Dataset (Dataset)

type Props
  = { dataset :: Dataset
    , defaultTemplate :: String
    , onNameChange :: Maybe String -> Effect Unit
    , onEnter :: String -> Effect Unit
    , onEsc :: Effect Unit
    }

type State
  = { snapshotTemplate :: String
    , snapshotName :: Maybe String
    , error :: Maybe String
    , showHelp :: Boolean
    }

data Action
  = ConvertTemplate String

update :: Self Props State -> Action -> Effect Unit
update self = case _ of

  ConvertTemplate s -> do
    self.setState _ { snapshotTemplate = s }
    ts <- nowDateTime
    either
      ( \error ->
          self.setState _ { error = Just error, snapshotName = Nothing }
            *> self.props.onNameChange Nothing
      )
      ( \name ->
          self.setState _ { snapshotName = Just name, error = Nothing }
            *> self.props.onNameChange (Just name)
      )
      (convert ts s >>= validateName)

snapshotNameForm :: Props -> JSX
snapshotNameForm = make component { initialState, didMount, render }
  where
  component :: Component Props
  component = createComponent "SnapshotNameForm"

  initialState = { snapshotTemplate: "", snapshotName: Nothing, error: Nothing, showHelp: false }

  didMount self =
    self.setState _ { snapshotTemplate = self.props.defaultTemplate }
      *> update self (ConvertTemplate self.props.defaultTemplate)

  render self =
    R.form
      { className: "mx-3"
      , onSubmit: capture_ $ pure unit
      , children:
        [ div "form-group row"
            $ R.label
                { htmlFor: "snapshot-name-template"
                , children: [ R.text "Snapshot name template" ]
                }
            <> R.input
                { className:
                  "form-control"
                    <> guard (isJust self.state.error) " is-invalid"
                , id: "snapshot-name-template"
                , autoFocus: true
                , placeholder: "Snapshot name template"
                , onChange: capture targetValue (fromMaybe "" >>> ConvertTemplate >>> update self)
                , onKeyDown: handler key $
                  case _ of
                    Just "Enter" -> foldMap self.props.onEnter self.state.snapshotName
                    Just "Escape" -> self.props.onEsc
                    _ -> pure unit
                , value: self.state.snapshotTemplate
                }
            <> foldMap (R.text >>> div "invalid-feedback") self.state.error
        , div "form-group"
            $ R.small
              { className: "form-text pointer text-primary"
              , onClick: capture_ $ self.setState \s -> s { showHelp = not s.showHelp }
              , children:
                let
                  txt =
                    if self.state.showHelp then
                      "Hide supported format sequences."
                    else
                      "Show supported format sequences."
                in
                 [ R.text txt ]
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
                , value: fromMaybe "" self.state.snapshotName
                }
        ]
      }

  div className child = R.div { className, children: [ child ] }

validateName :: String -> Either String String
validateName name =
  -- https://wiki.openindiana.org/oi/ZFS+naming+conventions
  let
    validStrs =
      (toEnumWithDefaults bottom top >>> S.singleton <$> 48 .. 57 <> 65 .. 90 <> 97 .. 122)
        <> [ "_", "-", ":", "." ]

    invalidStrs = A.filter (flip A.elem validStrs >>> not) $ (S.singleton <$> S.toCodePointArray name)
  in
   if S.null name then
     Left $ "Name can't be empty"
   else if A.null invalidStrs then
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
