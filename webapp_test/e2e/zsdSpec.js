describe('zsd', function(){
  beforeEach(function(){browser.get('http://localhost:12345')});
  
  it('should start with browse-actual', function(){
    expect(browser.getLocationAbsUrl()).toMatch(/browse-actual$/);
  });

});
