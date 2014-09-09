var zsd = angular.module('ZFSSnapDiff', ['ngRoute', 'ngSanitize']);

zsd.config(['$routeProvider', function($routeProvider){
  $routeProvider.
    when('/', {redirectTo: '/by-file'}).
    when('/by-file', {
      controller: 'ByFileCtrl as ctrl',
      templateUrl: 'by-file.html'
    }).
    when('/by-snapshots', {
      controller: 'BySnapshotCtrl as ctrl',
      templateUrl: 'by-snapshots.html' 
    }).
    otherwise({redirectTo: '/by-file'});
}]);


zsd.factory('Config', ["$http", function($http){
  var config;
  var promise = $http.get('config').then(function(res){
    config = res.data;
  });

  return {
    promise: promise,
    config: config,
    config: function(){
      return config;
    },
    get: function(key){
      return config[key]
    }
  };
}]);
zsd.factory('Backend', ["$http", "Config", function($http, Config){
  return {
    listSnapshots: function(whereFileModified){
      var params = {};
      if(whereFileModified !== undefined) params['where-file-modified'] = whereFileModified;
      
      return $http.get("list-snapshots", {"params": params}).then(function(res){
        return res.data
      });
    },
    listDir: function(path){
      return $http.get("list-dir", {"params": {"path": path}}).then(function(res){
        return res.data
      });
    },
    snapshotDiff: function(snapName){
      return $http.get("snapshot-diff", {"params": {"snapshot-name": snapName}}).then(function(res){
        return res.data
      })
    },
    readFile: function(path, snapName){
      var params = {"path": path, "max-file-size": Config.get("MaxFileSize")};
      if(snapName !== undefined) params['snapshot-name'] = snapName;
      
      return $http.get("read-file", {"params": params, "responseType": "arraybuffer"}).then(function(res){
        return {contentType: res.headers("Content-Type"),
                contentLength: res.headers("Content-Length"),
                isText: function(){ return res.headers("Content-Type").indexOf("text") >= 0},
                content: res.data};
      })
    }
  }
}]);

zsd.factory("Difflib", ["Config", function(Config){
  return {
    htmlDiff: function(currentFileContent, snapName, snapFileContent){
      var currentLines = difflib.stringAsLines(currentFileContent);
      var snapLines = difflib.stringAsLines(snapFileContent);
      
      var sm = new difflib.SequenceMatcher(snapLines, currentLines);

      return diffview.buildView({
        baseTextName: snapName,
        baseTextLines: snapLines,
        newTextName: "Current Version",
        newTextLines: currentLines,
        opcodes: sm.get_opcodes(),
        contextSize: Config.get("DiffContextSize"),
        viewType: 0 }).outerHTML;
    }
  }
}]);


zsd.directive('snapshots', [function(){
  return {
    restrict: 'E',
    templateUrl: 'template-snapshots.html',
    scope: {
      snapshots: '=',
      onSnapshotSelected: '&'
    },
    link: function($scope, $element, $attrs){
      $scope.snapshotSelected = function(snap){
        $scope.onSnapshotSelected({snap: snap});
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



zsd.directive('dirBrowser', ['Backend', function(Backend){
  return {
    restrict: 'E',
    templateUrl: 'template-dir-browser.html',
    scope: {
      startPath: '@',
      onFileSelected: '&',
      onDirSelected: '&'
    },
    link: function($scope, $element, $attrs){
      $scope.pathElements = [];
      $scope.fileSelected = false;

      $scope.filterHiddenEntries = function(entry){
        if(! $scope.showHiddenEntries){
          if(entry.Path) return entry.Path.charAt(0) != '.';
        }
        return true;
      };

      $scope.isDirectory = function(entry){
        return entry.Type === "D"
      };
      
      $scope.isFile = function(entry){
        return entry.Type === "F"
      };
      
      $scope.open = function(entry){
        var idx = $scope.pathElements.indexOf(entry);
        if(idx === -1){
          // user go deeper
          $scope.pathElements = $scope.pathElements.concat([entry]);
        }else{
          // user jump upward
          $scope.pathElements = $scope.pathElements.slice(0, idx + 1);
        }
        
        if($scope.isDirectory(entry)){
          $scope.dirEntries = [{}];
          $scope.fileSelected = false;
          $scope.onDirSelected({pathElements: $scope.pathElements});
          
          var path = $scope.pathElements.map(function(e){ return e.Path}).join("/")
          Backend.listDir(path).then(function(dirEntries){
            $scope.dirEntries = dirEntries;
          });
        }else{
          $scope.fileSelected = true;
          $scope.onFileSelected({pathElements: $scope.pathElements});
        }
      };

      // open at start element
      $scope.open({'Path': $scope.startPath, 'Type': 'D'});      
    }
  };
}]);

zsd.controller('MainCtrl', ["$location", "Config", function($location, Config){
  var self = this;


  Config.promise.then(function(){
    self.config = Config.config();
  });

  
  
  self.activeClassIfAt = function(path){
    return {active: $location.path() === path};
  };

}]);


zsd.controller('BySnapshotCtrl', ["Backend", "Difflib", "$timeout", "$window", function(Backend, Difflib, $timeout, $window){
  var self = this;

  Backend.listSnapshots().then(function(snapshots){
    self.snapshots = snapshots;
  });
  
  self.showSnapshotDiff = function(snap){
    self.currentSnapshot = snap;
    Backend.snapshotDiff(snap.Name).then(function(diff){
      self.snapshotDiff = diff;

      // delayed - to give the browser time
      $timeout(function(){
        $('#snapshotDiff')[0].scrollIntoView(true);
        $window.scrollBy(0, -80); // scroll up (FIXME: get from css: padding-top)
      }, 200);      
    })
  };

}]);


zsd.controller('ByFileCtrl', ["Backend", "Difflib", "$timeout", "$window", "$sce", "$q", function(Backend, Difflib, $timeout, $window, $sce, $q){
  var self = this;

  self.fileSelected = function(pathElements){
    delete self.snapshots;
    delete self.currentSnapshot
    delete self.fileDiff;
    delete self.binaryContent;
    self.isFileSelected = true
    var path = pathElements.map(function(e){ return e.Path}).join("/")
    self.currentPath = path;
    Backend.listSnapshots(path).then(function(s){
      self.snapshots = s;
    });   
  }

  self.snapshotSelected = function(snap, scrollToContent){
    self.currentSnapshot = snap;

    var fetchPromise = Backend.readFile(self.currentPath, snap.Name).then(function(snapFile){
      if(snapFile.isText()){
        return Backend.readFile(self.currentPath).then(function (currentFile){
          return {snapFile: snapFile, currentFile: currentFile};
        });
      }else{
        return $q.reject({status: "BINARY_CONTENT", file: snapFile});
      }
    });


    fetchPromise.then(function(res){
      var snapFileContent = arrayBuffer2String(res.snapFile.content);
      var currentFileContent = arrayBuffer2String(res.currentFile.content);     
      self.fileDiff = Difflib.htmlDiff(currentFileContent, snap.Name, snapFileContent);
    }, function(res){
      if(res.status === 406) {
        // max size exceeded -> trigger download
        $window.location = "/read-file?path="+self.currentPath+"&snapshot-name="+snap.Name;
      }

      if(res.status === "BINARY_CONTENT"){
        // show binary content
        var blob = new Blob([res.file.content], {type: res.file.contentType});
        var url = URL.createObjectURL(blob);
        self.binaryContent = $sce.trustAsResourceUrl(url);
      }      
    }).then(function(ignore){
      if(scrollToContent === true){
      // scroll to content
      $timeout(function(){
        $('#content')[0].scrollIntoView(true);
        $window.scrollBy(0, -80); // scroll up (FIXME: get from css: padding-top)
      }, 200);
      }
    });
  };


  self.currentSnapshotIsFirst = function(){
    return self.snapshots && self.snapshots.indexOf(self.currentSnapshot) === 0;
  };
  
  self.currentSnapshotIsLast = function(){
    return self.snapshots && self.snapshots.indexOf(self.currentSnapshot) === self.snapshots.length - 1;
  };
  
  self.previousSnapshot = function(){
    var idx = self.snapshots.indexOf(self.currentSnapshot);
    self.snapshotSelected(self.snapshots[idx - 1], false);
  };

  self.nextSnapshot = function(){
    var idx = self.snapshots.indexOf(self.currentSnapshot);
    self.snapshotSelected(self.snapshots[idx + 1], false);
  };

  self.downloadSnapFile = function(){
    $window.location = "/read-file?path="+self.currentPath+"&snapshot-name="+self.currentSnapshot.Name;
  }


  function arrayBuffer2String(arrayBuffer) {
    var str = '';
    var bytes = new Uint8Array(arrayBuffer);
    for(var i = 0, max = bytes.length; i < max; i++){
      str += String.fromCharCode(bytes[i]);
    }
    return str;
  }
  


}]);
