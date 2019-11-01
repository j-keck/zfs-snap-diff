angular.module('zsdUtils', ['zsdServices']).

// FileUtils
factory('FileUtils', [function(){
  var comparableMimeTypePrefixes = ["text"];
  var viewableMimeTypePrefixes = ["text", "image", "application/pdf"];

  return {
    isViewable: function(fileInfo){
      return viewableMimeTypePrefixes.filter(function(prefix){
        return fileInfo.mimeType.indexOf(prefix) >= 0;
      }).length > 0
    },
    isComparable: function(fileInfo){
      return comparableMimeTypePrefixes.filter(function(prefix){
        return fileInfo.mimeType.indexOf(prefix) >= 0;
      }).length > 0
    },
    isText: function(fileInfo){
      return fileInfo.mimeType.indexOf("text") >= 0;
    }
  }
}]).


// PathUtils
factory('PathUtils', ['Config', function(Config){
  // search the dataset which is the parent from the given path
  findDatasetForFile = function(path){
    var datasets = Config.get('datasets');

    // create a copy before sorting to keep the orginal dataset order intact
    datasets = datasets.slice(0);

    // sort the datasets - longest path at first
    datasets = datasets.sort(function(a, b){
      return b.mountPoint.path.length - a.mountPoint.path.length
    });

    for(var i in datasets){
      if(path.indexOf(datasets[i].mountPoint.path+"/") >= 0){
        return datasets[i]
      }
    }
  }
  return {
    convertToSnapPath: function(path, snapName){
      var dataset = findDatasetForFile(path);
      var relativePath = path.substring(dataset.mountPoint.path.length)
      return dataset.mountPoint.path + "/.zfs/snapshot/" + snapName + relativePath;
    },

    convertToActualPath: function(path){
      var dataset = findDatasetForFile(path)
      var mountPoint = dataset.mountPoint.path;
      var snapName = this.extractSnapName(path);

      var prefix = mountPoint + "/.zfs/snapshot/" + snapName;
      var relativePath = path.substring(prefix.length);

      return mountPoint + relativePath;
    },

    extractSnapName: function(path){
      var dataset = findDatasetForFile(path)
      var p = path.substring(dataset.mountPoint.path.length) // remove mount point
      p = p.substring('/.zfs/snapshot/'.length) // remove /.zfs/snapshot/
      var snapName = p.substring(0, p.indexOf('/')); // extract snapshot-name
      return snapName;
    },


      entriesToPath: function(entries){
          //return entries.map(function(e){ return e.Name}).join('/');
          return entries[entries.length - 1].path;
    },


    replaceRoot: function(entries, newRoot){
      var newEntries = entries;
      newEntries.shift(); // remove root
      newEntries.unshift(newRoot); // add new root
      return newEntries;
    },

    extractFileName: function(path){
      var pathElements = path.split('/');
      return pathElements[pathElements.length - 1];
    },


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
