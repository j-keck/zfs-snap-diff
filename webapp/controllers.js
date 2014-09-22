angular.module('zsdControllers', ['zsdServices', 'zsdUtils']).
  
controller('MainCtrl', ['$location', '$rootScope', '$timeout', 'Config', function($location, $rootScope, $timeout, Config){
  var self = this;
  
  Config.promise.then(function(){
    self.config = Config.config();
  });

  $rootScope.$on('zsd:error', function(event, msg){
    self.error = msg;
  });

  $rootScope.$on('zsd:warning', function(event, msg){
    self.warning = msg;
  });

  $rootScope.$on('zsd:success', function(event, msg){
    self.success = msg;
  });
  

  $rootScope.$on('zsd:http-activity', function(event, args){
    // first http request pending
    if(typeof self.timeoutHndl === 'undefined'){
      // delayed - only show spinner when duration > 'delay'
      var delay = 500;
      self.timeoutHndl = $timeout(function(){
        self.loading = true;
      }, delay);      
    }

    // no more http request pending
    if(args === 0){
      $timeout.cancel(self.timeoutHndl);
      delete self.timeoutHndl;
      self.loading = false;
    }
  });
  
  
  self.activeClassIfAt = function(path){
    return {active: $location.path() === path};
  };

}]).




controller('BrowseActualCtrl', ['Backend', 'PathUtils', 'Config', function(Backend, PathUtils, Config){
  var self = this;

  self.fileSelected = function(entries){
    delete self.curSnap;
    delete self.snapshots;

    var path = PathUtils.entriesToPath(entries);
    self.curPath = path;


    Backend.listSnapshots(
      path,
      Config.get('scanSnapLimit'),
      Config.get('compareFileMethod')
    ).then(function(snapshots){
      self.snapshots = snapshots;
    });   
  }

  
  self.dirSelected = function(entries){
    delete self.curSnap;
    delete self.curPath;    
    delete self.snapshots;
  }

  
  self.snapshotSelected = function(snap){
    self.curSnap = snap;
  };

}]).




controller('BrowseSnapshotsCtrl', ['Backend', 'PathUtils', function(Backend, PathUtils){
  var self = this;

  Backend.listSnapshots().then(function(snapshots){
    self.snapshots = snapshots;
  });


  self.snapshotSelected = function(snap){
    if(typeof self.curSnap === 'undefined'){
      // first time
      self.startEntries = [{Type: 'D', Path: snap.Path}];
    }else{
      // use last path - update only root element
      self.startEntries = PathUtils.replaceRoot(self.entries, {Type: 'D', Path: snap.Path});
    }
    self.curSnap = snap;
  };
  

  self.fileSelected = function(entries){
    self.entries = entries;
    var path = PathUtils.entriesToPath(entries);
    self.curPath = path;
  };

  self.dirSelected = function(entries){
    self.entries = entries;
    delete self.curPath;
  };

}]).




controller('BrowseSnapshotDiffCtrl', ['Backend', function(Backend){
  var self = this;

  Backend.listSnapshots().then(function(snapshots){
    self.snapshots = snapshots;
  });

  
  self.snapshotSelected = function(snap){
    self.curSnap = snap;
    delete self.snapshotDiff;
    Backend.snapshotDiff(snap.Name).then(function(diff){
      self.snapshotDiff = diff;
    });
  };
  
}]);
