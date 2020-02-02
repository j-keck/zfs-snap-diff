module ZSD.Utils.LocalStorage
       ( getItem
       , setItem
       ) where

import Prelude

import Data.Either (hush)
import Data.Maybe (Maybe)
import Effect (Effect)
import Effect.Console (log)
import Effect.Uncurried (EffectFn1, EffectFn2, runEffectFn1, runEffectFn2)
import Simple.JSON (class ReadForeign, class WriteForeign, readJSON, writeJSON)

foreign import setItem_ :: EffectFn2 String String Unit
foreign import getItem_ :: EffectFn1 String String

setItem :: forall a. WriteForeign a => String -> a -> Effect Unit
setItem k v = log "setItem" *> (runEffectFn2 setItem_ k $ writeJSON v)

getItem :: forall a. ReadForeign a => String -> Effect (Maybe a)
getItem k = do
  v <- runEffectFn1 getItem_ k
  pure $ (hush <<< readJSON) v
