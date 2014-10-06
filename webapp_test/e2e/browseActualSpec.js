describe('browse-actual', function(){
  beforeEach(function(){browser.get('http://localhost:12345')});  

  
  it('should have the file: file1 listed', function(){
    var entries = element.all(by.repeater('entry in dirListing'));
    filterByText(entries, "file1").then(function(filtered){
      expect(filtered.length).toEqual(1);
      expect(filtered[0].getText()).toMatch(/file1/);
    });
  });


  it('should list snapshots when click on a file', function(){
    var snapshots = function(){
      return element.all(by.repeater('snap in snapshots'))
    }

    expect(snapshots().count()).toMatch(0);
    clickOnEntry('file1');
    expect(snapshots().count()).toBeGreaterThan(0);
  });


  it('should show the file content when click on a snapshot', function(){
    var textFileContentIsPresentShouldBe = function(b){
      expect(element(by.binding('textFileContent')).isPresent()).toEqual(b)
    }

    textFileContentIsPresentShouldBe(false);
    
    clickOnEntry('file1');
    textFileContentIsPresentShouldBe(false);

    clickOnSnapshot('snap1');
    textFileContentIsPresentShouldBe(true);
  });


  it('should show the file diff when click on "Diff" button', function(){
    var diffIsPresentShouldBe = function(b){
      expect(element(by.repeater('diffContext in diffResult.intext')).isPresent()).toEqual(b)
    }

    diffIsPresentShouldBe(false);

    clickOnEntry('file1');
    diffIsPresentShouldBe(false);

    clickOnSnapshot('snap1');
    diffIsPresentShouldBe(false);

    // click on "Diff" button
    element(by.id('compareFile')).click();
    diffIsPresentShouldBe(true);
  });


  it('should show the diff from the child2 dataset when click on "Diff" button in the child2 dataset', function(){
    clickOnEntry('child2');
    clickOnEntry('file1');
    clickOnSnapshot('snap1');
    element(by.id('compareFile')).click();
    expect(element(by.binding('diffContext')).getText()).toMatch(/SECOND:child2/);
  });



  function filterByText(elements, s){
    return elements.filter(function(e){
      return e.getText().then(function(text){
        return text.indexOf(s) >= 0
      });
    });
  };


  function clickOnEntry(name){
    filterByText(element.all(by.repeater('entry in dirListing')), name).then(function(filtered){
      filtered[0].click();
    });    
  }

  function clickOnSnapshot(name){
    filterByText(element.all(by.repeater('snap in snapshots')), name).then(function(filtered){
      filtered[0].click();
    });
  }

});
