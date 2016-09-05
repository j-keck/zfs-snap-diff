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
  }).
  // filesize in b, kb, mb, gb from bytes
  filter('filesize', function(){
    var units = ['B', 'K', 'M', 'G', 'T', 'P'];
    return function(bytes){
      if(isFinite(bytes)){
        var i = 0;
        while(bytes >= 1024){
          bytes /= 1024;
          i++;
        }

        return bytes.toFixed(i >= 1 ? 1 : 0) + units[i];
      }else{
        return bytes;
      }
    }
  });
