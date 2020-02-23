-- | Simple component to scroll in the page
module ZSD.Components.Scroll where

import Effect (Effect)
import Prelude (Unit, (>>=))
import Web.HTML as H
import Web.HTML.Window as W

scroll :: Int -> Int -> Effect Unit
scroll x y = H.window >>= W.scroll x y

scrollToTop :: Effect Unit
scrollToTop = scroll 0 0
