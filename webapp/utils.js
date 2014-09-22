angular.module('zsdUtils', ['zsdServices']).

// FileUtils
factory('FileUtils', [function(){
  var comparableMimeTypePrefixes = ["text"];
  var viewableMimeTypePrefixes = ["text", "image", "application/pdf"];

  return {
    isViewable: function(fileInfo){
      return viewableMimeTypePrefixes.filter(function(prefix){
        return fileInfo.MimeType.indexOf(prefix) >= 0;
      }).length > 0
    },
    isComparable: function(fileInfo){
      return comparableMimeTypePrefixes.filter(function(prefix){
        return fileInfo.MimeType.indexOf(prefix) >= 0;
      }).length > 0
    },
    isText: function(fileInfo){
      return fileInfo.MimeType.indexOf("text") >= 0;
    }
  }
}]).


// PathUtils
factory('PathUtils', ['Config', function(Config){
  return {
    convertToSnapPath: function(path, snapName){
      var relativePath = path.substring(Config.get('ZFSMountPoint').length)
      return Config.get('ZFSMountPoint') + "/.zfs/snapshot/" + snapName + relativePath;
    },
    
    convertToActualPath: function(path){
      var mountPoint = Config.get('ZFSMountPoint');
      var snapName = this.extractSnapName(path);

      var prefix = mountPoint + "/.zfs/snapshot/" + snapName;
      var relativePath = path.substring(prefix.length);

      return mountPoint + relativePath;
    },

    extractSnapName: function(path){
      var pathElements = path.split('/');

      // remove mount point path-prefix
      for(var _ in Config.get('ZFSMountPoint').split('/')){
        pathElements.shift();
      }

      pathElements.shift(); // remove: .zfs
      pathElements.shift(); // remove: snapshots
      var snapName = pathElements.shift(); // snapName
      return snapName;
    },


    entriesToPath: function(entries){
      return entries.map(function(e){ return e.Path}).join('/');
    },


    entriesTargetsToFile: function(entries){
      return entries[entries.length - 1].Type === 'F';
    },


    replaceRoot: function(entries, newRoot){
      var newEntries = entries;
      newEntries.shift(); // remove root
      newEntries.unshift(newRoot); // add new root
      return newEntries;
    }

    
  }
}]).


factory('StringUtils', function(){
  return {
    trimPrefix: function(s, prefix){
      if(s.indexOf(prefix) === 0){
        return s.substring(prefix.length);
      }
      return s;
    }
  }
});
