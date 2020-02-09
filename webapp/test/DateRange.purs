module Test.DateRange where

import Prelude

import Test.Unit (TestSuite, suite, test)
import Test.Unit.QuickCheck (quickCheck)
import Test.QuickCheck (class Arbitrary, (===))
import Test.QuickCheck.Gen (chooseInt, choose, Gen)
import Data.Date as D
import Data.DateTime as DT
import Data.Time.Duration (Days(..))
import Data.Enum (toEnum, class BoundedEnum)
import Data.Maybe (fromMaybe)
import ZSD.Model.DateRange (DateRange(..))


tests :: TestSuite
tests = suite "DateRange" do
  test "semigroup" $ quickCheck
    \(ArbDateRange a) (ArbDateRange b) ->
      (a <> b) === (b <> a)


newtype ArbDateRange = ArbDateRange DateRange
instance showArgDateRange :: Show ArbDateRange where
  show (ArbDateRange dr) = show dr
instance arbDateRange :: Arbitrary ArbDateRange where
  arbitrary = do
    from <-     DT.canonicalDate
           <$> lift (chooseInt 1 31)
           <*> lift (chooseInt 1 12)
           <*> lift (chooseInt 1900 2050)
    to <- addDays <$> choose 0.1 100.0 <*> pure from
    pure <<< ArbDateRange <<< DateRange $ { from, to }

    where
      lift :: forall a. Bounded a => BoundedEnum a => Gen Int -> Gen a
      lift = map (fromMaybe bottom <<< toEnum)

      addDays n d = fromMaybe d $ D.adjust (Days n) d

