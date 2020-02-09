module Test.Main where

import Prelude

import Effect (Effect)
import Test.DateTime as DateTime
import Test.Unit.Main (runTest)
import Test.DateRange as DateRange

main :: Effect Unit
main = runTest do
  DateTime.tests
  DateRange.tests
