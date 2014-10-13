angular.module('zsdDirectives', ['zsdUtils', 'zsdServices']).


// https://github.com/angular/angular.js/issues/339
directive('zsdEmbedSrc', function () {
  return {
    restrict: 'A',
    link: function (scope, element, attrs) {
      var current = element;
      scope.$watch(function() { return attrs.zsdEmbedSrc; }, function () {
        var clone = element
          .clone()
          .attr('src', attrs.zsdEmbedSrc);
        current.replaceWith(clone);
        current = clone;
      });
    }
  };
}).


// zsd-show-if-defined is like ng-show but:
//  * shows content if 'angular.isDefined' returns true
//  * empty strings, lists or objects are defined
//    (not so with ng-show)
directive('zsdShowIfDefined', function(){
  return {
    restrict: 'A',
    link: function(scope, element, attrs){
      scope.$watch(function(){ return scope.$eval(attrs.zsdShowIfDefined)}, function(value){
        if(angular.isDefined(value)){
          attrs.$removeClass('hide');          
        }else{
          attrs.$addClass('hide');
        }
      });
    }
  }
}).


// zsd-show-if-empty is like ng-show but:
//  * shows content if 'angular.isDefined' returns true and value.length == 0
//     -> value is undefined: hide content
//     -> value is defined but empty: show content
//     -> value is defined and not empty: hide content
//  * usable for notifications like 'no xxx found'
directive('zsdShowIfEmpty', function(){
  return {
    restrict: 'A',
    link: function(scope, element, attrs){
      scope.$watch(function(){ return scope.$eval(attrs.zsdShowIfEmpty)}, function(value){
        if(angular.isDefined(value) && value.length == 0){
          attrs.$removeClass('hide');          
        }else{
          attrs.$addClass('hide');
        }
      });
    }
  }
}).


directive('zsdModal', [function(){
  return {
    restrict: 'E',
    scope: {
      show: '='
    },
    replace: true,
    transclude: true,
    link: function(scope, element, attrs) {
      scope.dialogStyle = {};
      if (attrs.width)
        scope.dialogStyle.width = attrs.width;
      if (attrs.height)
        scope.dialogStyle.height = attrs.height;
    },
    template: "<div class='zsd-modal' ng-show='show'>\n <div class='zsd-modal-overlay'></div>\n <div class='zsd-modal-dialog panel panel-default' ng-style='dialogStyle'>\n <div class='zsd-modal-dialog-content' ng-transclude></div>\n</div>\n</div>"
    
  }
}]).





// show notifications 
directive('zsdNotifications', ['$rootScope', '$timeout', 'Notifications', function($rootScope, $timeout, Notifications){
  return {
    restrict: 'E',
    scope: {
      // timeout after the message is removed
      timeout: '@'
    },
    link: function(scope, element, attrs){
      scope.messages = [];
      
      // remove message with the given id
      scope.removeMessage = function(id){
        scope.messages = scope.messages.filter(function(n){return n.id !== id});
      }

      // register a handler with receives new messages
      // from the Notification service
      Notifications.registerListener(function(msg){
        scope.messages.push(msg);
        // auto-remove old messages only when a timeout are given
        if(angular.isDefined(scope.timeout)){
          $timeout(function(){
            scope.removeMessage(msg.id);
          }, scope.timeout);
        }
      });      
    },
    templateUrl: "template-notifications.html"
  }
}]).
  
directive('zsdScrollToTop', ['$window', function($window){
  return function(scope, element, attrs){
    element.bind('click', function(event){
      $window.scrollTo(0, 0);
    });
  }
}]);
