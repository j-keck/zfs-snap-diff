describe('FileUtils', function(){

  var textFile = {MimeType: 'text/plain'};
  var pdfFile = {MimeType: 'application/pdf'};
  var videoFile = {MimeType: 'video/example'};

  beforeEach(module('ZFSSnapDiff'));

  // isViewable
  describe('isViewable', function(){
    it('should return true for text files', inject(function(FileUtils){
      expect(FileUtils.isViewable(textFile)).toEqual(true);
    }));

    it('should return true for pdf files', inject(function(FileUtils){
      expect(FileUtils.isViewable(pdfFile)).toEqual(true);
    }));
    
    it('should return false for video files', inject(function(FileUtils){
      expect(FileUtils.isViewable(videoFile)).toEqual(false);
    }));
  });


  // isComparable
  describe('isComparable', function(){
    it('should return true for text files', inject(function(FileUtils){
      expect(FileUtils.isComparable(textFile)).toEqual(true);
    }));

    it('should return false for pdf files', inject(function(FileUtils){
      expect(FileUtils.isComparable(pdfFile)).toEqual(false);
    }));

    it('should return false for video files', inject(function(FileUtils){
      expect(FileUtils.isComparable(videoFile)).toEqual(false);
    }));
  });

  
});
