angular.module('zsdDirBrowser', ['zsdServices']). 
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
          return entry.Type === "DIR"
        };

        scope.isFile = function(entry){
          return entry.Type === "FILE"
        };

        // returns 'true' if the given entry is not a regular file / directory
        scope.isSpecialEntry = function(entry){
          return ! (scope.isFile(entry) || scope.isDirectory(entry));
        };

        // returns the bootstrap icon class for the given entry
        scope.iconClassForEntry = function(entry) {
          switch(entry.Type) {
          case "FILE": return "glyphicon-file";
          case "DIR": return "glyphicon-folder-open";
          case "LINK": return "glyphicon-link";
          case "PIPE": return "glyphicon-transfer";
          case "SOCKET": return "glyphicon-transfer";
          case "DEV": return "glyphicon-hdd";
          };
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



        scope.$watch(function(){ return scope.start}, function(start){
          if(! angular.isDefined(start)) return;
          
          scope.entries = [];
          scope.open({Type: 'DIR', Path: scope.start});
        });
        
        scope.$watch(function(){ return scope.startEntries}, function(startEntries){
          if(! angular.isDefined(startEntries)) return;
          scope.entries = startEntries;

          // start on last element
          scope.open(scope.entries[scope.entries.length - 1]);
        });

      }
    };
  }]);

