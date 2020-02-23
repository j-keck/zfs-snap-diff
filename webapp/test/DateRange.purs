module Test.DateRange where

import Prelude

import Data.Date as D
import Data.DateTime as DT
import Data.Enum (toEnum, class BoundedEnum)
import Data.JSDate as JSDate
import Data.Maybe (fromMaybe)
import Data.Time.Duration (Days(..))
import Effect.Class (liftEffect)
import Test.QuickCheck (class Arbitrary, (===))
import Test.QuickCheck.Gen (chooseInt, choose, Gen)
import Test.Unit (TestSuite, suite, test)
import Test.Unit.Assert as Assert
import Test.Unit.QuickCheck (quickCheck)
import ZSD.Model.DateRange (DateRange(..))
import ZSD.Model.DateRange as DateRange
import ZSD.Utils.Ops (unsafeFromJust)


tests :: TestSuite
tests = suite "DateRange" do
  test "last0Days" do
    dr <- liftEffect $ DateRange.lastNDays (Days 0.0)
    Assert.equal 1 $ DateRange.dayCount dr

  test "last1Days" do
    dr <- liftEffect $ DateRange.lastNDays (Days 1.0)
    Assert.equal 2 $ DateRange.dayCount dr

  test "last2Days" do
    dr <- liftEffect $ DateRange.lastNDays (Days 2.0)
    Assert.equal 3 $ DateRange.dayCount dr


  test "days" do
    let unsafeDate = JSDate.parse
                     >>> map (JSDate.toDate >>> unsafeFromJust)
                     >>> liftEffect
    from <- unsafeDate "2020-01-01"
    to   <- unsafeDate "2020-01-02"
    let dr = DateRange { from, to }
    Assert.equal 2 $ DateRange.dayCount dr

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
