angular.module('zsdControllers', ['zsdServices', 'zsdUtils']).

controller('MainCtrl', ['$location', '$rootScope', '$timeout', 'Config', function($location, $rootScope, $timeout, Config){
  var self = this;

  Config.promise.then(function(){
    self.config = Config.config();
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




controller('BrowseActualCtrl', ['Backend', 'PathUtils', 'Config', 'Session', function(Backend, PathUtils, Config, Session){
  var self = this;

  self.datasetSelected = function(dataset){
    delete self.curSnap;
    delete self.curPath;
    delete self.snapshots;

    self.curDataset = dataset;
    Session.set('curDataset', dataset);
  }

  self.fileSelected = function(entries){
    delete self.curSnap;
    delete self.snapshots;

    var path = PathUtils.entriesToPath(entries);
    self.curPath = path;
    self.curFileName = PathUtils.extractFileName(path);


    Backend.snapshotsForFile(
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


  // start on 'curDataset' if it's defined
  if(Session.has('curDataset')){
    self.datasetSelected(Session.get('curDataset'))
  }

}]).




controller('BrowseSnapshotsCtrl', ['Backend', 'PathUtils', 'Session', function(Backend, PathUtils, Session){
  var self = this;

  self.datasetSelected = function(dataset){
    self.curDataset = dataset;
    Session.set('curDataset', dataset);

    delete self.curSnap;
    delete self.curPath;

    Backend.snapshotsForDataset(dataset.name).then(function(snapshots){
      self.snapshots = snapshots;
    });
  }


  self.snapshotSelected = function(snap){
    if(typeof self.curSnap === 'undefined'){
      // first time
      self.startEntries = [{kind: 'DIR', path: snap.path}];
    }else{
      // use last path - update only root element
      self.startEntries = PathUtils.replaceRoot(self.entries, {kind: 'DIR', path: snap.path});
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

  // start on 'curDataset' if it's defined
  if(Session.has('curDataset')){
    self.datasetSelected(Session.get('curDataset'))
  }
}]).




controller('BrowseSnapshotDiffCtrl', ['Backend', 'Session', function(Backend, Session){
  var self = this;

  self.datasetSelected = function(dataset){
    self.curDataset = dataset;
    Session.set('curDataset', dataset);

    delete self.snapshotDiff;
    Backend.snapshotsForDataset(dataset.name).then(function(snapshots){
      self.snapshots = snapshots;
    });
  }



  self.snapshotSelected = function(snap){
    self.curSnap = snap;
    delete self.snapshotDiff;
    Backend.snapshotDiff(self.curDataset.name, snap.name).then(function(diff){
      self.snapshotDiff = diff;
    });
  };

  // start on 'curDataset' if it's defined
  if(Session.has('curDataset')){
    self.datasetSelected(Session.get('curDataset'))
  }

}]).

controller('BrowseMessagesCtrl', ['Notifications', function(Notifications){
  var self = this;
  self.messages = Notifications.messages();
}]);
