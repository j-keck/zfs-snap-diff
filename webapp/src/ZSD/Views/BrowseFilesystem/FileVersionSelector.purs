module ZSD.Views.BrowseFilesystem.FileVersionSelector where

import Prelude

import Data.Array as A
import Data.Either (either)
import Data.Foldable (foldMap)
import Data.Int (toNumber)
import Data.Maybe (Maybe(..), maybe)
import Data.Monoid (guard)
import Data.Newtype (unwrap)
import Data.Time.Duration (Days(..))
import Data.Tuple (Tuple(..))
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Foreign.Object as O
import React.Basic (Component, JSX, createComponent, fragment, make, readState)
import React.Basic as React
import React.Basic.DOM as R
import React.Basic.DOM.Events (capture_)
import React.Basic.DOM.Textf as TF
import ZSD.Components.DropDownButton (dropDownButton)
import ZSD.Components.Panel (panel)
import ZSD.Components.Spinner as Spinner
import ZSD.Components.TableX (tableX)
import ZSD.Model.DateRange (DateRange(..))
import ZSD.Model.DateRange as DateRange
import ZSD.Model.FH (FH)
import ZSD.Model.FileVersion (FileVersion(..), ScanResult(..), scanBackups)
import ZSD.Utils.Formatter as Formatter
import ZSD.Utils.Ops (foldlSemigroup)
import ZSD.Views.Messages as Messages



type Props =
  { file :: FH
  , onVersionSelected :: FileVersion -> Effect Unit
  }

type State = { versions :: Array FileVersion
             , selectedIdx :: Int
             , scanResults :: Array ScanResult
             , scanDays :: Int
             }

data Action =
    DidMount
  | Scan Days Action
  | SelectVersionByIdx Int
  | NoOp


data SelectedVersion =
    Actual FileVersion
  | Backup FileVersion Int

update :: React.Self Props State -> Action -> Effect Unit
update self = case _ of
  DidMount -> do
    self.setState _ { versions = [ ActualVersion self.props.file ] }
    update self $ Scan (Days $ -1.0) (SelectVersionByIdx 0)

  Scan days next -> Spinner.display *> do
    range <- readState self >>= \state ->
             maybe (DateRange.lastNDays days)
                   (unwrap >>> _.dateRange >>> DateRange.slide days >>> pure)
                   (A.last state.scanResults )
    launchAff_ $ do
      res <- scanBackups self.props.file range
      liftEffect $ either Messages.appError (\scanResult ->  do
        state <- readState self
        let versions = A.concat [state.versions, (unwrap scanResult).fileVersions]
        self.setStateThen (const $ state { scanResults = state.scanResults `A.snoc` scanResult
                                         , versions = versions
                                         })
                                 $ update self next *> Spinner.remove) res


  SelectVersionByIdx idx -> do
    state <- readState self
    case A.index state.versions idx of
      Just next -> self.setStateThen _ { selectedIdx = idx } $ self.props.onVersionSelected next
      Nothing -> guard (hasOlderSnapshots state) $
                        update self $ Scan (Days $ -1.0) (SelectVersionByIdx idx)

  NoOp -> pure unit


fileVersionSelector :: Props -> JSX
fileVersionSelector = make component { initialState, didMount, render }

  where

     component :: Component Props
     component = createComponent "FileVersionSelector"

     initialState = { versions: [], selectedIdx: 0 , scanResults: [], scanDays: 1 }

     didMount self = update self DidMount

     render self = fragment
       [ panel
         { header: fragment
           [ R.text $ "Versions for file: " <> (unwrap self.props.file).name
           , R.span
             { className: "float-right"
             , children:
               [ R.div
                 { className: "btn-group"
                 , children:
                   [ R.button
                     { className: "btn btn-secondary" <> guard  (not $ hasOlderVersions self.state) " disabled"
                     , title: "Select the previous version"
                     , onClick: capture_ $ guard (hasOlderVersions self.state)
                                         $ update self (SelectVersionByIdx (self.state.selectedIdx + 1)) 
                     , children:
                       [ R.span { className: "fas fa-backward p-1" }
                       , R.text "Older"
                       ]
                     }
                   , R.button
                     { className: "btn btn-secondary" <> guard (not $ hasNewerVersions self.state) " disabled"
                     , title: "Select the successor version"
                     , onClick: capture_ $ guard (hasNewerVersions self.state)
                                         $ update self (SelectVersionByIdx (self.state.selectedIdx - 1))
                     , children:
                       [ R.text "Newer"
                       , R.span { className: "fas fa-forward p-1" }
                       ]
                     }
                   ]
                 }
               ]
             }
           ]
         , body: \hidePanelBodyFn ->
             fragment
             [ tableX
               { header: ["File modification time", "Snapshot Created", "Snapshot Name"]
               , rows: self.state.versions
               , mkRow: case _ of
                   ActualVersion _     -> [ R.text $ "Actual version", R.text "-", R.text "-" ]
                   BackupVersion {backup, snapshot} ->
                     [ R.text $ Formatter.dateTime (unwrap backup).mtime
                     , R.text $ Formatter.dateTime snapshot.created
                     , R.text snapshot.name
                     ]
               , onRowSelected: \(Tuple idx v) -> do
                   hidePanelBodyFn
                   self.setState _ { selectedIdx = idx }
                   self.props.onVersionSelected v
               , activeIdx: Just self.state.selectedIdx
               }
               
             , dropDownButton
               { content: R.text $ "Scan " <> show self.state.scanDays <> " days back"
               , title: "Scan for other file versions for " <> show self.state.scanDays <> " days on the server"
               , disabled: not $ hasOlderSnapshots self.state
               , onClick: update self $ Scan (Days $ negate (toNumber self.state.scanDays)) NoOp
               , entries:
                     let forDays n = Tuple (R.text $ show n) (self.setState _ { scanDays = n })
                     in map forDays [1, 7, 14, 30, 60, 180 ]
               , entriesTitle: "Change how many days should be scanned"
               }
             ]

         , footer:
             R.div
             { className: "text-muted small text-center"
             , children:
               [ flip foldMap (A.last self.state.scanResults)
                   \(ScanResult { snapsScanned, dateRange: (DateRange range), scanDuration }) ->
                     R.div_
                     [ TF.textf
                       [ TF.text { style: TF.textUnderline } "This scan: "
                       , TF.text' "Scanned ", TF.int { style: TF.fontBolder } snapsScanned, TF.text' " snapshots in "
                       , TF.text { style: TF.fontBolder } (Formatter.duration scanDuration), TF.text' ". "
                       , TF.text' "Date range: "
                       , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.from, TF.text' " - "
                       , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.to, TF.text' "."
                       ]
                     ]
               , flip foldMap (foldlSemigroup self.state.scanResults)
                   \(ScanResult { snapsScanned, snapsToScan, snapsFileMissing, dateRange: (DateRange range), scanDuration }) ->
                     R.div_
                     [ TF.textf
                       [ TF.text { style: TF.textUnderline } "Overall: "
                       , TF.text' "Scanned ", TF.int { style: TF.fontBolder } snapsScanned, TF.text' " snapshots in "
                       , TF.text { style: TF.fontBolder } (Formatter.duration scanDuration), TF.text' ". "
                       , TF.text' "Date range: "
                       , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.from, TF.text' " - "
                       , TF.date { format: [TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy] } range.to, TF.text' "."
                       ]
                    ] <>
                    R.div_
                    [ TF.textf
                      [ TF.int { style: TF.fontBolder } snapsToScan, TF.text' " snapshots to scan."
                      , TF.text' " In ", TF.int { style: TF.fontBolder } snapsFileMissing, TF.text' " snapshots was the file not found."
                      ]
                    ]
                ]
             }
         }
       ]






hasNewerVersions :: State -> Boolean
hasNewerVersions state = state.selectedIdx /= 0

hasOlderVersions :: State -> Boolean
hasOlderVersions state = state.selectedIdx < (A.length state.versions) - 1 ||
                         hasOlderSnapshots state

hasOlderSnapshots :: State -> Boolean
hasOlderSnapshots state = maybe false (unwrap >>> _.snapsToScan >>> (/=) 0) (A.last state.scanResults)


fileWasMissingInLastScan :: State -> Boolean
fileWasMissingInLastScan state = maybe false (unwrap >>> _.snapsFileMissing >>> (/=) 0) (A.last state.scanResults)
