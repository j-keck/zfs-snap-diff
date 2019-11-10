module ZSD.Components.FileVersionSelector where

import Prelude
import Data.Array.NonEmpty as ANE
import Data.Either (fromRight)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Console (log, logShow)
import Effect.Ref as Ref
import Effect.Unsafe (unsafePerformEffect)
import Partial.Unsafe (unsafePartial)
import React.Basic (Component, JSX, createComponent, make)
import React.Basic.DOM as R
import React.Basic.DOM.Components.LogLifecycles (logLifecycles)
import ZSD.Component.Table (table)
import ZSD.Components.Panel (panel)
import ZSD.Formatter as Formatter
import ZSD.Model.FSEntry (FSEntry)
import ZSD.Model.FileVersion (FileVersion(..), FileVersions)
import ZSD.Model.FileVersion as FileVersions



type Props =
  { file :: FSEntry
  , onVersionSelected :: FileVersion -> Effect Unit
  }

type State = { versions :: FileVersions }

fileVersionSelector :: Props -> JSX
fileVersionSelector props = logLifecycles $ make component { initialState, didMount, render } props

  where

     component :: Component Props
     component = createComponent "FileVersionSelector"

     initialState = { versions: ANE.singleton (ActualVersion props.file) }

     didMount self = launchAff_ $ do
       -- FIXME: error
       versions <- unsafePartial $ fromRight <$> FileVersions.fetch self.props.file
       liftEffect $ self.setState \s -> s { versions = ANE.appendArray s.versions versions }

     render self =
       let showPanelBody = unsafePerformEffect $ Ref.new true in
       panel
       { title: "Versions for file: " <> self.props.file.name
       , body:
           table
           { header: ["File modification time", "Snapshot Created", "Snapshot Name"]
             , rows: ANE.toArray self.state.versions
             , mkRow: case _ of
                 ActualVersion f -> [ R.text $ "Actual version", R.text "-", R.text "-" ]
                 BackupVersion v -> [ R.text $ Formatter.dateTime v.file.modTime
                                   , R.text $ Formatter.dateTime v.snapshot.created
                                   , R.text v.snapshot.name
                                   ]
             , onRowSelected: \v -> do
                 Ref.write false showPanelBody
                 self.props.onVersionSelected v
             }
       , showBody: showPanelBody
       }
