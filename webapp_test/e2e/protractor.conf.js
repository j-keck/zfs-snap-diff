exports.config = {
  seleniumAddress: 'http://localhost:4444/wd/hub',
  specs: ['zsdSpec.js', 'browseActualSpec.js'],
  capabilities: {
    //  'browserName': 'firefox'
    'browserName': 'phantomjs'
  }
}
