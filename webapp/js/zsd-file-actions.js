// 
// zsd-file-actions directive
//
// Usage:
//   from browsing acutal fs:
//     <zsd-file-actions zsd-from-actual
//                       path="path-in-actual" ....
//
//   from browsing in snapshot:
//     <zsd-file-actions zsd-from-snapshot
//                       path="path-in-snap"...
//
angular.module('zsdFileActions', ['zsdServices', 'zsdUtils']).
  directive('zsdFileActions', ['$window', '$sce', '$rootScope', 'FileUtils', 'Backend', 'PathUtils', 'Config', function($window, $sce, $rootScope, FileUtils, Backend, PathUtils, Config){
    return {
      restrict: 'E',
      scope: true,
      templateUrl: 'template-file-actions.html',      
      controller: function($scope){

        // 
        // controller actions 
        // 
               
        // trigger actions if a file is selected
        this.fileSelected = function(pathInActual, pathInSnap, curSnap){
          $scope.pathInActual = pathInActual;
          $scope.pathInSnap = pathInSnap;
          $scope.curSnap = curSnap;
          
          // fetch file-info
          Backend.fileInfo(pathInSnap).then(function(fi){
            $scope.fileInfo = fi;

            // for ui: enable / disable view and diff buttons
            $scope.fileIsViewable = FileUtils.isViewable(fi);
            $scope.fileIsComparable = FileUtils.isComparable(fi);

            // trigger lastAction
            $scope.lastAction();
          });
        };


        

        // 
        // scope (ui) actions
        // 
        
        // view the file content from the selected snapshot
        $scope.viewFile = function viewFile(){
          $scope.lastAction = $scope.viewFile;
          
          if(FileUtils.isText($scope.fileInfo)){
            Backend.readTextFile($scope.pathInSnap).then(function(res){
              clearOthersButKeep('textFileContent');
              
              // apply syntax highlight
              var hljsRes = hljs.highlightAuto(res);
              $scope.textFileContent = hljsRes.value;
            });
          }else{
            Backend.readBinaryFile($scope.pathInSnap).then(function(res){
              clearOthersButKeep('binaryFileContent');
              
              var url = URL.createObjectURL(res);
              $scope.binaryFileContent = $sce.trustAsResourceUrl(url);              
            });
          }
        }

        // compare the file content from the selected snapshot with the actual state
        $scope.compareFile = function compareFile(){
          $scope.lastAction = $scope.compareFile;
          Backend.diffFile($scope.pathInActual, $scope.curSnap.Name).then(function(res){
            clearOthersButKeep('diffResult');      
            $scope.diffResult = res;          
          })
        };

        // download the file from the selected snapshot
        $scope.downloadFile = function downloadFile(){
          $window.location = "/read-file?path=" + $scope.pathInSnap;
        };


        // show restore confirmation
        $scope.restoreFile = function restoreFile(){
          $scope.showRestoreFileConfirmation = true;
        };

        // restore the file from the selected snapshot
        $scope.restoreFileAcked = function(){
          $scope.hideRestoreFileConfirmation();

          Backend.restoreFile($scope.pathInActual, $scope.curSnap.Name).then(function(res){
            $rootScope.$broadcast('zsd:success', res);
            $scope.lastAction()
          });
        };

        // hide restore confirmation
        $scope.hideRestoreFileConfirmation = function(){
          delete $scope.showRestoreFileConfirmation;
        };
        
        // returns 'active' if a given name equals the function name from the lastAction
        //   * for action buttons 'toggle'
        $scope.activeClassIfLastActionIs = function(name){
          if($scope.lastAction.name === name){
            return "active";
          }
        };


        //
        // initializations
        //   * initialize lastAction
        //     - needs be after function declarations
        // 

        // initialize lastAction to a default value from the config
        var actions = {'off': function(){},
                       'view': $scope.viewFile,
                       'diff': $scope.compareFile,
                       'download': $scope.downloadFile,
                       'restore': $scope.restoreFile};
        
        var defaultFileAction = Config.get('defaultFileAction');
        if(defaultFileAction in actions){
          $scope.lastAction = actions[defaultFileAction];
        }else{
          $root$Scope.$broadcast('zsd:warning', 'Invalid "default-file-action": "'+ defaultFileAction +'"');
          $scope.lastAction = actions['off'];
        }
        

        
        // 
        // private / helper actions
        // 

        // clear other content, but keep the content with the given name
        function clearOthersButKeep(keep){
          if(keep !== 'diffResult')
            delete $scope.diffResult;

          if(keep !== 'textFileContent')
            delete $scope.textFileContent;

          if(keep !== 'binaryFileContent')
            delete $scope.binaryFileContent;
        }
      }
    }
  }]).


directive('zsdFromActual', ['PathUtils', function(PathUtils){
  return {
    restrict: 'A',
    require: '^zsdFileActions',
    scope:{
      path: "=",
      curSnap: "="
    },    
    link: function(scope, element, attrs, ctrl){
      // 
      // initializations
      //   * register observers for:
      //     - path changes (user change the current file)
      //     - snapshot changes (user switch between file versions)
      //

      // watch for path changes
      scope.$watch('path', function(path){
        scope.pathInActual = path;
        scope.pathInSnap = PathUtils.convertToSnapPath(path, scope.curSnap.Name);
        
        // trigger fileSelected
        ctrl.fileSelected(scope.pathInActual, scope.pathInSnap, scope.curSnap);
      });

      // watch for new snapshot selected
      scope.$watch('curSnap', function(snap){
        // update path vars (pathInActual doesn't change when browsing in the history)
        scope.pathInSnap = PathUtils.convertToSnapPath(scope.pathInActual, snap.Name)

        // trigger fileSelected
        ctrl.fileSelected(scope.pathInActual, scope.pathInSnap, snap);
      });
    }
  }
}]).

directive('zsdFromSnapshot', ['PathUtils', function(PathUtils){
  return {
    restrict: 'A',
    require: '^zsdFileActions',
    scope: {
      path: "=",
      curSnap: "=",
    },
    link: function(scope, element, attrs, ctrl){
      // 
      // initializations
      //   * register observers for:
      //     - path changes (user change the current file / switch between file versions)
      //       path starts with the snapshot mount-point, no need to observe snapshot changes

      // watch for path changes
      scope.$watch('path', function(path){

        // update path vars
        scope.pathInActual = PathUtils.convertToActualPath(path);
        scope.pathInSnap = path;
        
        // trigger fileSelected
        ctrl.fileSelected(scope.pathInActual, scope.pathInSnap, scope.curSnap);
      });
    }
  }
}]);
