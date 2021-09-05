module ZSD.Views.BrowseFilesystem.Stats where

import Prelude
import Data.Array as A
import Data.Traversable (foldMap)
import React.Basic (JSX)
import React.Basic.Classic (createComponent, makeStateless)
import React.Basic.DOM as R
import React.Basic.DOM.Textf as TF
import ZSD.Model.DateRange (DateRange(..))
import ZSD.Model.FileVersion (FileVersion, ScanResult(..))
import ZSD.Utils.Formatter as Formatter
import ZSD.Utils.Ops (foldlSemigroup)

type Props
  = { scanResults :: Array ScanResult
    , versions :: Array FileVersion
    }

stats :: Props -> JSX
stats =
  makeStateless component \props ->
    flip foldMap (foldlSemigroup props.scanResults)
      $ \(ScanResult { snapsScanned, snapsToScan, scanDuration, dateRange: (DateRange { from, to }) }) ->
          R.div
            { className: "text-muted small text-center"
            , children:
              A.singleton
                $ TF.textf
                    [ TF.text' "Scanned "
                    , TF.int { style: TF.fontBold } snapsScanned
                    , TF.text' " from "
                    , TF.int { style: TF.fontBold } (snapsScanned + snapsToScan)
                    , TF.text' " snapshots in "
                    , TF.text { style: TF.fontBold } (Formatter.duration scanDuration)
                    , TF.text' "."
                    , TF.jsx (R.br {})
                    , TF.int { style: TF.fontBold } (A.length props.versions)
                    , TF.text' " file versions between "
                    , datef from
                    , TF.text' " and "
                    , datef to
                    , TF.text' " found."
                    ]
            }
  where
  component = createComponent "Stats"

  datef = TF.date { format: [ TF.dd, TF.s " ", TF.mmmm, TF.s " ", TF.yyyy ], style: TF.fontBold }
