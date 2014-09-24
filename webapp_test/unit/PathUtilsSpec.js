describe('PathUtils', function(){

  var zfsMountPoint = "/base/path";
  var relativePath  = "/relative/path/to/file";

  var testSnapName = "20140001-1d";
  var testPathInActual = zfsMountPoint + relativePath
  var testPathInSnapshot = zfsMountPoint + "/.zfs/snapshot/" + testSnapName + relativePath;


  
  beforeEach(module('ZFSSnapDiff'));

  beforeEach(module(function($provide){
    $provide.value('Config', {
      get: function(key){
        return zfsMountPoint;
      }
    });
  }));


  describe('convertToSnapPath', function(){
    it('should convert a path in actual to a path under a given snapshot', inject(function(PathUtils){
      var pathInSnapshot = PathUtils.convertToSnapPath(testPathInActual, testSnapName);
      expect(pathInSnapshot).toEqual(testPathInSnapshot);
    }));
  });

  describe('convertToLivePath', function(){
    it('should convert a path in a snapshot to a path under actual', inject(function(PathUtils){
      var pathInActual = PathUtils.convertToActualPath(testPathInSnapshot);
      expect(pathInActual).toEqual(testPathInActual);
    }));    
  });

  describe('extractSnapName', function(){
    it('should extract snapshot name from a path', inject(function(PathUtils){
      var snapName = PathUtils.extractSnapName(testPathInSnapshot);
      expect(snapName).toEqual(testSnapName)
    }));
  });


  describe('entriesToPath', function(){
    it('should return the path as string', inject(function(PathUtils){
      var entries = [{Path: 'home'}, {Path: 'user'}, {Path: 'filename'}];
      expect(PathUtils.entriesToPath(entries)).toEqual('home/user/filename');
    }))
  });

  describe('entriesTargetsToFile', function(){
    it('should return true if the last element i a file', inject(function(PathUtils){
      var entries = [{Type: 'D'}, {Type: 'D'}, {Type: 'F'}];
      expect(PathUtils.entriesTargetsToFile(entries)).toEqual(true);
    }));

    it('should return false if the last element i a dir', inject(function(PathUtils){
      var entries = [{Type: 'D'}, {Type: 'D'}, {Type: 'D'}];
      expect(PathUtils.entriesTargetsToFile(entries)).toEqual(false);
    }));
  });


  describe('replaceRoot', function(){
    it('should replace the root element', inject(function(PathUtils){
      var entries = [{Path: 'a'}, {Path: 'b'}, {Path: 'c'}];
      var expected = [{Path: 'A'}, {Path: 'b'}, {Path: 'c'}];
      expect(PathUtils.replaceRoot(entries, {Path: 'A'})).toEqual(expected);
    }));
  });
  
  
});
