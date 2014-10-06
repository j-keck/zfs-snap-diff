describe('PathUtils', function(){

  var datasets = [
    {Name: 'zp', MountPoint: '/zp'},
    {Name: 'zp/a', MountPoint: '/zp/a'},
    {Nmae: 'zp/a/b', MountPoint: '/zp/a/b'},
  ];
  
  var zfsMountPoint = "/zp/a";
  var relativePath  = "/relative/path/to/file";

  var testSnapName = "20140001-1d";
  var testPathInActual = zfsMountPoint + relativePath
  var testPathInSnapshot = zfsMountPoint + "/.zfs/snapshot/" + testSnapName + relativePath;


  
  beforeEach(module('ZFSSnapDiff'));

  beforeEach(module(function($provide){
    $provide.value('Config', {
      get: function(key){
        return datasets;
      }
    });
  }));


  describe('convertToSnapPath', function(){
    it('should convert a path in actual to a path under a given snapshot', inject(function(PathUtils){
      var pathInSnapshot = PathUtils.convertToSnapPath(testPathInActual, testSnapName);
      expect(pathInSnapshot).toEqual(testPathInSnapshot);
    }));
  });

  describe('convertToActualPath', function(){
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


  describe('replaceRoot', function(){
    it('should replace the root element', inject(function(PathUtils){
      var entries = [{Path: 'a'}, {Path: 'b'}, {Path: 'c'}];
      var expected = [{Path: 'A'}, {Path: 'b'}, {Path: 'c'}];
      expect(PathUtils.replaceRoot(entries, {Path: 'A'})).toEqual(expected);
    }));
  });
  
  
});
