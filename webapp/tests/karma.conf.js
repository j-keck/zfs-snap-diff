module.exports = function(config){
  config.set({
    basePath: '',
    frameworks: ['jasmine'],
    files: [
      '../lib/angular/angular.min.js',
      '../lib/angular/angular-mocks.js',
      '../lib/angular/angular-route.min.js',
      '../lib/angular/angular-sanitize.min.js',
      '../lib/angular/angular-animate.min.js',
      '../*.js',
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
