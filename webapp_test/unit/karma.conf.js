module.exports = function(config){
  config.set({
    basePath: '',
    frameworks: ['jasmine'],
    files: [
      '../../webapp/lib/angular/angular.min.js',
      '../../webapp/lib/angular/angular-mocks.js',
      '../../webapp/lib/angular/angular-route.min.js',
      '../../webapp/lib/angular/angular-sanitize.min.js',
      '../../webapp/lib/angular/angular-animate.min.js',
      '../../webapp/js/*.js',
      '*.js'
    ],
    exclude: [],
    port: 8082,
    logLevel: config.LOG_INFO,
    autoWatch: true,
    browsers: ['PhantomJS'],
    singleRun: false
  })
};  
