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
      // chancel if server is unavailable
      if(rejection.status === 0){
        $rootScope.$broadcast('zsd:error', 'Server unavailable');
        return $q.reject(rejection);
      }
      

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


factory('Config', ['$http', function($http){
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
  
factory('Backend', ['$http', 'Config', function($http, Config){
  return {
    snapshotsForFile: function(path, scanSnapLimit, compareFileMethod){
      var params = {path: path};

      if(angular.isDefined(scanSnapLimit)) params['scan-snap-limit'] = scanSnapLimit;
      if(angular.isDefined(compareFileMethod)) params['compare-file-method'] = compareFileMethod;
      
      return $http.get('snapshots-for-file', {'params': params}).then(function(res){
        return res.data
      });
    },
    snapshotsForDataset: function(name){
      return $http.get('snapshots-for-dataset', {'params': {'dataset-name': name}}).then(function(res){
        return res.data;
      });
    },
    listDir: function(path){
      return $http.get('list-dir', {'params': {'path': path}}).then(function(res){
        return res.data
      });
    },
    snapshotDiff: function(datasetName, snapName){
      return $http.get('snapshot-diff', {'params': {'dataset-name': datasetName, 'snapshot-name': snapName}}).then(function(res){
        return res.data
      })
    },
    readTextFile: function(path){
      return this.readFile(path, 'text')
    },
    readBinaryFile: function(path){
      return this.readFile(path, 'blob')
    },
    readFile: function(path, responseType){
      var params = {path: path};
    
      return $http.get('read-file', {'params': params, 'responseType': responseType}).then(function(res){
        return res.data;
      });
    },
    fileInfo: function(path){
      // fileInfo with caching
      return $http.get('file-info', {'params': {'path': path}, 'cache': true}).then(function(res){
        return res.data;
      })
    },
    restoreFile: function(path, snapName){
      return $http.put('restore-file', {'path': path, 'snapshot-name': snapName}).then(function(res){
        return res.data;
      });
    },

    diffFile: function(path,snapName){
      return $http.get('diff-file', {params: {path: path, 'snapshot-name': snapName, 'context-size': Config.get('diffContextSize')}}).then(function(res){
        return res.data;
      });
    },

    revertChange: function(path, deltas){
      return $http.put('revert-change', {path: path, deltas: deltas}).then(function(res){
        return res.data;
      });
    }

  }
}]).

factory('Notifications', ['$rootScope', function($rootScope){
  var messages = [];
  var listeners = [];
  
  $rootScope.$on('zsd:error', function(event, msg){
    addMessage('error', msg);
  });

  $rootScope.$on('zsd:warning', function(event, msg){
    addMessage('warning', msg);
  });

  $rootScope.$on('zsd:success', function(event, msg){
    addMessage('success', msg);
  });


  // save a new message and notify listeners
  function addMessage(type, text){
    var id = messages.length + 1;
    var message = {id: id, ts: new Date(), type: type, text: text};

    messages.push(message);

    // notify listeners
    for(var i in listeners){
      listeners[i](message);
    }
  };

  return {
    registerListener: function(listener){
      listeners.push(listener);
    },

    deleteMessages: function(){
      messages = [];
    },

    messages: function(){
      return messages;
    }

    
  }
  
}]).

// session store
factory('Session', [function(){
  var store = {};
  return {
    set: function(key, value){
      store[key] = value;
    },
    get: function(key){
      return store[key];
    },
    has: function(key){
      return angular.isDefined(this.get(key));
    }
  }
}]);
