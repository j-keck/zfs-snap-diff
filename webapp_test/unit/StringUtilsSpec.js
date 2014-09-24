describe('StringUtils', function(){

  beforeEach(module('ZFSSnapDiff'));

  // trimPrefix
  describe('trimPrefix', function(){
    it('should return the whole string if it doesnt start with the prefix', inject(function(StringUtils){
      var string = "sample content";
      var result = StringUtils.trimPrefix(string, "content");
      expect(result).toEqual(string);
    }))


    it('should return the string without the prefix', inject(function(StringUtils){
      var string = "sample content";
      var result = StringUtils.trimPrefix(string, "sample ");
      expect(result).toEqual("content");
    }))
  });


             
});
