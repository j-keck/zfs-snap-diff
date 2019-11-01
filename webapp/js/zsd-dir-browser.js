angular.module('zsdDirBrowser', ['zsdServices']).
  directive('zsdDirBrowser', ['Backend', 'PathUtils', '$filter', function(Backend, PathUtils, $filter){
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
        scope.orderByPropName = "name";
        scope.orderReversed = false;

        scope.filterHiddenEntries = function(entry){
          if(! scope.showHiddenEntries){
            if(entry.name) return entry.name.charAt(0) != '.';
          }
          return true;
        };

        scope.isDirectory = function(entry){
          return entry.kind === "DIR"
        };

        scope.isFile = function(entry){
          return entry.kind === "FILE"
        };

        // returns 'true' if the given entry is not a regular file / directory
        scope.isSpecialEntry = function(entry){
          return ! (scope.isFile(entry) || scope.isDirectory(entry));
        };

        // returns the bootstrap icon class for the given entry
        scope.iconClassForEntry = function(entry) {
          switch(entry.kind) {
          case "FILE": return "glyphicon-file";
          case "DIR": return "glyphicon-folder-open";
          case "LINK": return "glyphicon-link";
          case "PIPE": return "glyphicon-transfer";
          case "SOCKET": return "glyphicon-transfer";
          case "DEV": return "glyphicon-hdd";
          };
        };

        scope.orderBy = function(propName, reversed) {
          // if parameter 'reversed' not given, toggle the value
          if(typeof reversed === "undefined"){
            scope.orderReversed = scope.orderByPropName === propName ? !scope.orderReversed : false;
          } else {
            scope.orderReversed = reversed;
          };
          scope.orderByPropName = propName;

          var orderBy = $filter("orderBy");

          if(propName === "name"){
            // split folders and files
            var [folders, files] = [Array(), Array()];
            var length = scope.dirListing.length;
            for(var i = 0; i < length; i++){
              if(scope.dirListing[i].kind === "DIR"){
                folders.push(scope.dirListing[i]);
              }else{
                files.push(scope.dirListing[i]);
              }
            }

            // sort the folders and files
            folders = orderBy(folders, "path", scope.orderReversed);
            files = orderBy(files, "path", scope.orderReversed);

            // concat the folders and files
            scope.dirListing = folders.concat(files);

          }else{
            var mapping = {"size": "size", "mtime": "modTime"};
            scope.dirListing = orderBy(scope.dirListing, mapping[propName], scope.orderReversed);
          }
        };

        scope.cssClassForCol = function(name){
          var icon, iconAlt;
          switch(name){
          case "name": [icon, iconAlt] = ["glyphicon-sort-by-alphabet", "glyphicon-sort-by-alphabet-alt"]; break;
          case "size": [icon, iconAlt] = ["glyphicon-sort-by-attributes", "glyphicon-sort-by-attributes-alt"]; break;
          case "mtime": [icon, iconAlt] = ["glyphicon-sort-by-attributes", "glyphicon-sort-by-attributes-alt"]; break;
          }

          return (scope.orderByPropName === name) ? "glyphicon " + (scope.orderReversed ? iconAlt : icon) : "";
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
              scope.orderBy(scope.orderByPropName, scope.orderReversed);
            });
          }else{
            scope.fileSelected = true;
            scope.onFileSelected({entries: scope.entries});
          }
        };



        scope.$watch(function(){ return scope.start}, function(start){
          if(! angular.isDefined(start)) return;

          scope.entries = [];
          scope.open({kind: 'DIR', path: scope.start});
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
