module Main where

import Prelude
import Data.Either (either)
import Data.Maybe (maybe)
import Effect (Effect)
import Effect.Aff (launchAff_)
import Effect.Class (liftEffect)
import Effect.Exception (throw)
import React.Basic.DOM (render)
import Web.DOM.NonElementParentNode (getElementById)
import Web.HTML (window)
import Web.HTML.HTMLDocument (toNonElementParentNode)
import Web.HTML.Window (document)
import ZSD.Model.Config as Config
import ZSD.WebApp (webApp)


main :: Effect Unit
main = do
  element <- lookupElement >>= maybe (throw "node with id 'webapp' not found") pure
  launchAff_ $ do
    config <- Config.fetch >>= either (show >>> throw >>> liftEffect) pure
    liftEffect $ render (webApp { config }) element

  where
    lookupElement = getElementById "webapp" =<< (map toNonElementParentNode $ document =<< window)
