describe('FileUtils', function(){
  
  var files = {
    "text-file": {MimeType: "text/plain"},
    "pdf-file": {MimeType: "application/pdf"},
    "video-file": {MimeType: "video/example"}
  };

  var deferred, rootScope;

  beforeEach(function () {
    module('ZFSSnapDiff', function($provide){
      $provide.value('Backend', {
        fileInfo: function(path){
          deferred.resolve(files[path]);
          return deferred.promise;
        }
      });
    });

    inject(function($rootScope, $q){
      rootScope = $rootScope;
      deferred = $q.defer();
    });
  });



  // isViewable
  describe('isViewable', function(){
    it('should return true for text files', inject(function(FileUtils){
      FileUtils.isViewable("text-file").then(function(res){
        expect(res).toEqual(true);
      });
      // propagate promise resolution
      rootScope.$apply();          
    }));

    it('should return true for pdf files', inject(function(FileUtils){
      FileUtils.isViewable("pdf-file").then(function(res){
        expect(res).toEqual(true);
      });
      rootScope.$apply();    
    }));
    
    it('should return false for video files', inject(function(FileUtils){
      FileUtils.isViewable("video-file").then(function(res){
        expect(res).toEqual(false);
      });
      rootScope.$apply();    
    }));
  });


  // whenIsViewable
  describe('whenIsViewable', function(){
    it('should execute the given function when its a viewable file', inject(function(FileUtils){
      var executed = false;
      FileUtils.whenIsViewable("text-file", function(){
        executed = true;
      });
      rootScope.$apply();
      expect(executed).toEqual(true);
    }));

    it('should NOT execute the given function when its no viewable file', inject(function(FileUtils){
      var executed = false;
      FileUtils.whenIsViewable("video-file", function(){
        executed = true;
      });
      rootScope.$apply();
      expect(executed).toEqual(false);
    }));
  });



  // isComparable
  describe('isComparable', function(){
    it('should return true for text files', inject(function(FileUtils){
      FileUtils.isComparable("text-file").then(function(res){
        expect(res).toEqual(true);
      });
      rootScope.$apply();    
    }));

    it('should return false for pdf files', inject(function(FileUtils){
      FileUtils.isComparable("pdf-file").then(function(res){
        expect(res).toEqual(false);
      });
      rootScope.$apply();    
    }));

    it('should return false for video files', inject(function(FileUtils){
      FileUtils.isComparable("video-file").then(function(res){
        expect(res).toEqual(false);
      });
      rootScope.$apply();    
    }));
  });

  // whenIsComparable
  describe('whenIsComparable', function(){
    it('should execute the given function when its a comparable file', inject(function(FileUtils){
      var executed = false;
      FileUtils.whenIsComparable("text-file", function(){
        executed = true;
      });
      rootScope.$apply();
      expect(executed).toEqual(true);
    }));

    it('should NOT execute the given function when its no comparable file', inject(function(FileUtils){
      var executed = false;
      FileUtils.whenIsComparable("video-file", function(){
        executed = true;
      });
      rootScope.$apply();
      expect(executed).toEqual(false);
    }));
  });
  
  
});
