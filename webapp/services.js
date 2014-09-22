angular.module('zsdServices', []).

factory('HTTPActivityInterceptor', ['$q', '$rootScope', '$timeout', function($q, $rootScope, $timeout){
  var activityCounter = 0;
  var timeoutHandlers = [];
  return {
    'request': function(config){
      activityCounter++;
      $rootScope.$broadcast('zsd:http-activity', activityCounter);

      return config
    },
    'response': function(response){
      activityCounter--;
      $rootScope.$broadcast('zsd:http-activity', activityCounter);

      return response;
    },
    'responseError': function(rejection){
      activityCounter--;
      $rootScope.$broadcast('zsd:http-activity', activityCounter);
      
      return $q.reject(rejection);      
    }
  };
}]).

factory('HTTPErrorInterceptor', ['$q', '$rootScope', function($q, $rootScope){
  return {
    'responseError': function(rejection){


      if(rejection.config.responseType === 'blob'){
        // convert message to string if result type is a blob and broadcast the error
        var reader = new FileReader();
        reader.readAsBinaryString(rejection.data);
        reader.onloadend = function(){
          $rootScope.$broadcast('zsd:error', reader.result);
        }
      }else{
        // already text - broadcast the error
        $rootScope.$broadcast('zsd:error', rejection.data);
      }
      return $q.reject(rejection);                      
    }
  };
}]).


factory('Config', ["$http", function($http){
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
}]).
  
factory('Backend', ["$http", "Config", function($http, Config){
  return {
    listSnapshots: function(whereFileModified, scanSnapLimit){
      var params = {};

      if(angular.isDefined(whereFileModified)) params['where-file-modified'] = whereFileModified;
      if(angular.isDefined(scanSnapLimit)) params['scan-snap-limit'] = scanSnapLimit;
      
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
    readTextFile: function(path){
      return this.readFile(path, "text")
    },
    readBinaryFile: function(path){
      return this.readFile(path, "blob")
    },
    readFile: function(path, responseType){
      var params = {path: path};
    
      return $http.get("read-file", {"params": params, "responseType": responseType}).then(function(res){
        return res.data;
      });
    },
    fileInfo: function(path){
      // fileInfo with caching
      return $http.get("file-info", {"params": {"path": path}, "cache": true}).then(function(res){
        return res.data;
      })
    },
    restoreFile: function(path, snapName){
      return $http.put("restore-file", {'path': path, 'snapshot-name': snapName}).then(function(res){
        return res.data;
      });
    }
  }
}]).


factory('Difflib', ['Config', 'Backend', function(Config, Backend){
  return {
    diffText: function(actualContent, snapName, snapContent){
      var actualLines = difflib.stringAsLines(actualContent);
      var snapLines = difflib.stringAsLines(snapContent);
      
      var sm = new difflib.SequenceMatcher(snapLines, actualLines);

      return diffview.buildView({
        baseTextName: snapName,
        baseTextLines: snapLines,
        newTextName: "Actual Version",
        newTextLines: actualLines,
        opcodes: sm.get_opcodes(),
        contextSize: Config.get("diffContextSize"),
        viewType: 0 }).outerHTML;
    },

    diffFiles: function(actualFile, snapName, snapFile){
      var self = this;
      var p = Backend.readTextFile(actualFile).then(function(actualContent){
        return Backend.readTextFile(snapFile).then(function(snapContent){
          return {actualContent: actualContent, snapContent: snapContent};
        })
      });

      return p.then(function(r){
        return self.diffText(r.actualContent, snapName, r.snapContent);
      });
    }
  }
}]);
