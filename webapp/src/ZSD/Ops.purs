-- | Module `ZSD.Ops` contains missing Purescript operators / functions
module ZSD.Ops where

import Data.Array as A
import Data.Tuple (Tuple(..))
import Prelude (class Applicative, class Bind, class Functor, map, (-), (<$>), (<*>), (>>>))


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
