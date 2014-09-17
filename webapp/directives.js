
zsd.directive('fileActions', ['$window', '$sce', 'FileUtils', 'Backend', 'Difflib', 'PathUtils', function($window, $sce, FileUtils, Backend, Difflib, PathUtils){
  return {
    restrict: 'E',
    templateUrl: 'template-file-actions.html',
    scope:{
      path: "=",
      pathFrom: "@",
      curSnap: "="
    },
    link: function(scope, element, attrs){
      scope.viewFile = function(){
        scope.lastAction = scope.viewFile;
        delete scope.fileDiff;
        delete scope.binaryFileContent;

        if(! scope.fileIsViewable) return;

        var path = realPath();
        FileUtils.isText(path).then(function(isText){
          if(isText){
            Backend.readTextFile(path).then(function(res){
              scope.textFileContent = res;
            });
          }else{
            Backend.readBinaryFile(path).then(function(res){
              var url = URL.createObjectURL(res);
              scope.binaryFileContent = $sce.trustAsResourceUrl(url);              
            });
          }
        });
      }
      
      scope.compareFile = function(){
        scope.lastAction = scope.compareFile;
        delete scope.textFileContent;
        delete scope.binaryFileContent;

        if(! scope.fileIsComparable) return;

        var actualPath, snapPath;
        if(scope.pathFrom === 'actual'){
          actualPath = scope.path;
          snapPath = PathUtils.convertToSnapPath(actualPath, scope.curSnap.Name);
        }else if(scope.pathFrom === 'snapshot'){
          snapPath = scope.path;
          actualPath = PathUtils.convertToActualPath(snapPath);
        }else{
          throw 'Invalid "path-from": ' + scope.pathFrom;
        }
        
        Difflib.diffFiles(actualPath, scope.curSnap.Name, snapPath).then(function(diff){
          scope.fileDiff = diff;
        });
      };
        
      scope.downloadFile = function(){
        delete scope.lastAction;
        $window.location = "/read-file?path="+realPath();
      };
      
      scope.restoreFile = function(){
        delete scope.lastAction;
      };

      
      scope.$watch('path', function(){
        if(typeof scope.path === 'undefined') return;

        // clear old state
        delete scope.fileDiff;
        delete scope.textFileContent;
        delete scope.binaryFileContent;


        FileUtils.isViewable(scope.path).then(function(res){
          scope.fileIsViewable = res;
        });
        
        FileUtils.isComparable(scope.path).then(function(res){
          scope.fileIsComparable = res;
        });

      });

      scope.$watch('curSnap', function(){
        if(typeof scope.curSnap === 'undefined') return;

        if(typeof scope.lastAction === 'undefined') return;
        scope.lastAction();
      });

      function realPath(){
        var path;
        if(scope.pathFrom === 'actual'){
          path = PathUtils.convertToSnapPath(scope.path, scope.curSnap.Name);
        }else if(scope.pathFrom === 'snapshot'){
          path = scope.path;
        }else{
          throw 'Invalid "path-from": ' + scope.pathFrom;
        }
        return path;
      }
    }
  }
}]);





zsd.directive('snapshots', ['$location', '$anchorScroll', function($location, $anchorScroll){
  return {
    restrict: 'E',
    templateUrl: 'template-snapshots.html',
    scope: {
      snapshots: '=',
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
}]);






// https://github.com/angular/angular.js/issues/339
zsd.directive('embedSrc', function () {
  return {
    restrict: 'A',
    link: function (scope, element, attrs) {
      var current = element;
      scope.$watch(function() { return attrs.embedSrc; }, function () {
        var clone = element
          .clone()
          .attr('src', attrs.embedSrc);
        current.replaceWith(clone);
        current = clone;
      });
    }
  };
});





zsd.directive('dirBrowser', ['Backend', 'PathUtils', function(Backend, PathUtils){
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
}]);


