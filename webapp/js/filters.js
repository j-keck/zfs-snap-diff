angular.module('zsdFilters', []).
  filter('ordinal', function(){
    return function(n){
      // code from: http://ecommerce.shopify.com/c/ecommerce-design/t/ordinal-number-in-javascript-1st-2nd-3rd-4th-29259
      var s = ["th", "st", "nd", "rd"],
      v = n % 100;
      return n + (s[(v-20)%10]||s[v]||s[0]);
    }
  }).
  // map 'error' text to 'danger' for bootstrap labels / alerts
  filter('mapErrorToDanger', function(){
    return function(text){
      if(text === 'error') return 'danger';
      return text;
    }
  });
