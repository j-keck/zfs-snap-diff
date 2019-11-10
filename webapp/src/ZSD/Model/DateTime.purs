module ZSD.Model.DateTime where

import Data.DateTime as DT
import Data.JSDate as JSDate
import Data.Maybe (fromJust)
import Data.Newtype (class Newtype)
import Effect.Unsafe (unsafePerformEffect)
import Foreign (readString)
import Partial.Unsafe (unsafePartial)
import Prelude (class Eq, class Show, map, ($), (<$>), (<<<))
import Simple.JSON (class ReadForeign)

newtype DateTime = DateTime DT.DateTime
derive newtype instance showDateTime :: Show DateTime
derive newtype instance eqDateTime :: Eq DateTime
derive instance newtypeDateTime :: Newtype DateTime _

instance readForeignModTime :: ReadForeign DateTime where
  readImpl f = DateTime <<< toDateTime' <$> readJSDate f
    where
      readJSDate = map (unsafePerformEffect <<< JSDate.parse) <<< readString
      toDateTime' = unsafePartial $ fromJust <<< JSDate.toDateTime
