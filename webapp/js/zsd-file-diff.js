angular.module('zsdFileDiff', ['zsdServices']).
  directive('zsdFileDiff', ['Backend', function(Backend){
    return {
      restrict: 'E',
      scope: {
        diffResult: '=',
        path: '=',
        fileName: '=',
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
