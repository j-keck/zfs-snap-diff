module Test.DateTime where

import Prelude
import Control.Monad.Free (Free)
import Data.Either (Either(..))
import Data.DateTime as DT
import Data.Enum (toEnum, class BoundedEnum)
import Data.Maybe (fromMaybe)
import Test.QuickCheck (class Arbitrary, (===))
import Test.QuickCheck.Gen (chooseInt, Gen)
import Test.Unit (TestF, suite, test)
import Test.Unit.QuickCheck (quickCheck)
import Simple.JSON (readJSON, writeJSON)
import ZSD.Model.DateTime (DateTime(..))


tests :: Free TestF Unit
tests = suite "date-time" do
  test "writeJSON / readJSON" $ quickCheck \(ArbDateTime a) ->
    Right a === (writeJSON >>> readJSON) a


newtype ArbDateTime = ArbDateTime DateTime
instance arbDateTime :: Arbitrary ArbDateTime where
  arbitrary = do
    date <-      DT.canonicalDate
            <$> lift (chooseInt 1 31)
            <*> lift (chooseInt 1 12)
            <*> lift (chooseInt 1900 2050)
    time <-      DT.Time
            <$> lift (chooseInt 0 23)
            <*> lift (chooseInt 0 59)
            <*> lift (chooseInt 0 59)
            <*> lift (chooseInt 0 999)
    pure $ ArbDateTime $ DateTime $ DT.DateTime date time

    where lift :: forall a. Bounded a => BoundedEnum a => Gen Int -> Gen a
          lift = map (fromMaybe bottom <<< toEnum)
