module ZSD.Model.DateTime where

import Prelude

import Data.DateTime as DT
import Data.JSDate as JSDate
import Data.Maybe (fromJust)
import Data.Newtype (class Newtype)
import Effect.Unsafe (unsafePerformEffect)
import Foreign (readString, unsafeToForeign)
import Partial.Unsafe (unsafePartial)
import Simple.JSON (class ReadForeign, class WriteForeign)


newtype DateTime = DateTime DT.DateTime


derive newtype instance showDateTime :: Show DateTime
derive newtype instance eqDateTime :: Eq DateTime
derive instance newtypeDateTime :: Newtype DateTime _
derive newtype instance ordDateTime :: Ord DateTime
instance boundedDateTime :: Bounded DateTime where
  top = DateTime top
  bottom = DateTime bottom

instance readForeignDateTime :: ReadForeign DateTime where
  readImpl f = DateTime <<< toDateTime <$> readJSDate f
    where
      readJSDate = map (unsafePerformEffect <<< JSDate.parse) <<< readString
      toDateTime = unsafePartial $ fromJust <<< JSDate.toDateTime

instance writeForeignDateTime :: WriteForeign DateTime where
  writeImpl (DateTime dt) = unsafeToForeign $ JSDate.fromDateTime dt

