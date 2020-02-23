-- | Module `ZSD.Ops` contains missing Purescript operators / functions
module ZSD.Utils.Ops where

import Prelude
import Data.Array as A
import Data.Either (Either, fromRight)
import Data.Foldable (class Foldable)
import Data.Foldable as F
import Data.Maybe (Maybe(..), fromJust, fromMaybe, maybe)
import Data.String as S
import Data.Tuple (Tuple(..))
import Partial.Unsafe (unsafePartial)

mapmap :: forall f1 f2 a b. Functor f1 => Functor f2 => (a -> b) -> f1 (f2 a) -> f1 (f2 b)
mapmap = map >>> map

infix 4 mapmap as <$$>

mapmapmap :: forall f1 f2 f3 a b. Functor f1 => Functor f2 => Functor f3 => (a -> b) -> f1 (f2 (f3 a)) -> f1 (f2 (f3 b))
mapmapmap = map >>> map >>> map

infix 4 mapmapmap as <$$$>

tupleM :: forall f a b. Applicative f => Bind f => f a -> f b -> f (Tuple a b)
tupleM a b = Tuple <$> a <*> b

zipWithIndex :: forall a. Array a -> Array (Tuple Int a)
zipWithIndex as = A.zipWith Tuple (A.range 0 (A.length as - 1)) as

foldlSemigroup :: forall a t. Foldable t => Semigroup a => t a -> Maybe a
foldlSemigroup = F.foldl (\b a -> maybe (Just a) (\b' -> Just $ b' <> a) b) Nothing

foldrSemigroup :: forall a t. Foldable t => Semigroup a => t a -> Maybe a
foldrSemigroup = F.foldr (\a b -> maybe (Just a) (\b' -> Just $ a <> b') b) Nothing

unsafeFromJust :: forall a. Maybe a -> a
unsafeFromJust a = unsafePartial $ fromJust a

unsafeFromRight :: forall a b. Either a b -> b
unsafeFromRight e = unsafePartial $ fromRight e

checkAll :: forall a f. F.Foldable f => f (a -> Boolean) -> a -> Boolean
checkAll fs a = F.foldl (\b f -> f a && b) true fs

checkAny :: forall a f. F.Foldable f => f (a -> Boolean) -> a -> Boolean
checkAny fs a = F.foldl (\b f -> f a || b) false fs

pathAppend :: String -> String -> String
pathAppend a b =
  let
    a' = fromMaybe a $ S.stripSuffix (S.Pattern "/") a

    b' = fromMaybe b $ S.stripPrefix (S.Pattern "/") b
  in
    a' <> "/" <> b'

infix 4 pathAppend as </>
