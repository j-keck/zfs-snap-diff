angular.module('zsdDirectives', ['zsdUtils', 'zsdServices']).

directive('zsdSnapshots', ['$location', '$anchorScroll', function($location, $anchorScroll){
  return {
    restrict: 'E',
    templateUrl: 'template-snapshots.html',
    scope: {
      snapshots: '=',
      title: '@',
      onSnapshotSelected: '&'
    },
    link: function(scope, element, attrs){
      
      scope.snapshotSelected = function(snap){
        scope.hideSnapshots = true;
        scope.curSnap = snap;
        scope.onSnapshotSelected({snap: snap});

        // scroll to top: FIXME:
        /*
        scope.$on('$locationChangeStart', function(ev) {
          ev.preventDefault();
        });
        $location.hash('top');
        $anchorScroll();
        */
      };
      
      scope.toggleHideSnapshots = function(){
        scope.hideSnapshots = ! scope.hideSnapshots;
      };

      scope.showNewerSnapDisabled = function(){
        return snapUninitialized() || scope.snapshots.indexOf(scope.curSnap) === 0
      };
      
      scope.showOlderSnapDisabled = function(){
        return snapUninitialized() || scope.snapshots.indexOf(scope.curSnap) === scope.snapshots.length - 1;
      };
      
      scope.showOlderSnap = function(){
        var idx = scope.snapshots.indexOf(scope.curSnap);
        scope.snapshotSelected(scope.snapshots[idx + 1]);
      };

      scope.showNewerSnap = function(){
        var idx = scope.snapshots.indexOf(scope.curSnap);
        scope.snapshotSelected(scope.snapshots[idx - 1]);
      };

      scope.$watch('snapshots', function(){
        // new file selected
        scope.hideSnapshots = false;
      });

      function snapUninitialized(){
        return typeof scope.curSnap === 'undefined' || typeof scope.snapshots === 'undefined';
      }
    }
  };
}]).






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


directive('zsdDirBrowser', ['Backend', 'PathUtils', function(Backend, PathUtils){
  return {
    restrict: 'E',
    templateUrl: 'template-dir-browser.html',
    scope: {
      start: '=',
      startEntries: '=',
      onFileSelected: '&',
      onDirSelected: '&'
    },
    link: function(scope, element, attrs){
      scope.fileSelected = false;

      scope.filterHiddenEntries = function(entry){
        if(! scope.showHiddenEntries){
          if(entry.Path) return entry.Path.charAt(0) != '.';
        }
        return true;
      };

      scope.isDirectory = function(entry){
        return entry.Type === "D"
      };
      
      scope.isFile = function(entry){
        return entry.Type === "F"
      };
      
      scope.open = function(entry){
        var idx = scope.entries.indexOf(entry);
        if(idx === -1){
          // user go deeper
          scope.entries = scope.entries.concat([entry]);
        }else{
          // user jump upward
          scope.entries = scope.entries.slice(0, idx + 1);
        }

        
        if(scope.isDirectory(entry)){
          scope.dirEntries = [{}];
          scope.fileSelected = false;
          scope.onDirSelected({entries: scope.entries});

          var path = PathUtils.entriesToPath(scope.entries);
          Backend.listDir(path).then(function(dirListing){
            scope.dirListing = dirListing;
          });
        }else{
          scope.fileSelected = true;
          scope.onFileSelected({entries: scope.entries});
        }
      };



      if(typeof scope.start !== 'undefined'){
        scope.entries = [];
        scope.open({Type: 'D', Path: scope.start});
      }
      
      scope.$watch(function(){ return scope.startEntries}, function(){
        if(typeof scope.startEntries === 'undefined') return;
        scope.entries = scope.startEntries;

        // start on last element
        scope.open(scope.entries[scope.entries.length - 1]);
      });

    }
  };
}]).

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





directive('zsdFileDiff', ['Backend', function(Backend){
  return {
    restrict: 'E',
    scope: {
      diffResult: '=',
      path: '=',
      curSnap: '='
    },
    templateUrl: 'template-file-diff.html',
    link: function(scope, element, attrs){
      // 
      // scope actions
      // 
      
      scope.showRevertChangeConfirmation = function(idx){
        scope.revertChangeConfirmation = idx;
      };
      scope.showRevertChangeConfirmationFor = function(idx){
        return  scope.revertChangeConfirmation == idx;
      };

      scope.downloadPatch = function(idx){
        var patch = unescape(scope.diffResult.patches[idx]);
        var patchName = PathUtils.extractFileName(scope.pathInActual) + ".patch";
          
        var link = $window.document.createElement('a');
        link.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(patch));
        link.setAttribute('download', patchName);
        link.click();
      };

      scope.revertChange = function(idx){
        delete scope.revertChangeConfirmation;
        
        Backend.revertChange(scope.path, scope.diffResult.deltas[idx]).then(function(res){
          Backend.diffFile(scope.path, scope.curSnap.Name).then(function(res){
            scope.diffResult = res;
          })
        })
      };

      // returns 'active' if a given name equals the current diffType
      //   * for diff type tabs
      scope.activeClassIfDiffTypeIs = function(name){
        if(scope.diffType === name){
          return "active";
        }
      };

      


      //
      // initializations      
      //
      scope.$watch('curSnap', function(){
        delete scope.revertChangeConfirmation;
      });
      
    }
  }
}]).

directive('zsdSideBySideDiffRows', ['$compile', function($compile){
  return {
    restrict: 'A',
    transclude: 'element',
    scope: {
      blocks: '='
    },
    compile: function(element, attr, linker){
      return function($scope, $element, $attr) {
        var childScopes = [];
        
        var parent = $element.parent();
        var header = parent.children();
        $scope.$watchCollection('blocks', function(blocks){
          for(var i in childScopes){
            childScopes[i].$destroy();
          }
          parent.html('');
          parent.append(header);

          //FIXME: cleanup!!!
          // delegate to downloadPatch from the caller side
          $scope.downloadPatch = function(idx){
            $scope.$parent.downloadPatch(idx);
          }

          //FIXME: cleanup!!!
          // delegate to revertChange from the caller side
          $scope.revertChange = function(idx){
            $scope.$parent.revertChange(idx);
          }

          
          // add new
          for(i in blocks){
            // create a new scope
            var childScope = $scope.$new();

            // pass patch counter as zsdIndex in the scope
            childScope['zsdIndex'] = +i;

            // FIXME: remove $scope.$parent ($emit / $broadcast?)
            // pass showRevertChangeConfirmation / showRevertChangeConfirmationFor in the scope
            childScope['showRevertChangeConfirmation'] = $scope.$parent.showRevertChangeConfirmation;
            childScope['showRevertChangeConfirmationFor'] = $scope.$parent.showRevertChangeConfirmationFor;


            linker(childScope, function(clone){
              // add to the DOM
              parent.append(clone);
              parent.append(blocks[i]);

              childScopes.push(childScope);
            });
          }
        });
      }
    }
  }
}]);
