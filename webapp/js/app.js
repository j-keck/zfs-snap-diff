var zsd = angular.module('ZFSSnapDiff',
                         ['ngRoute', 'ngSanitize', 'ngAnimate',
                          'zsdControllers', 'zsdUtils', 'zsdServices', 'zsdDirectives', 'zsdFileActions', 'zsdFilters', 'zsdDatasets']);

zsd.config(['$routeProvider', function($routeProvider){
  $routeProvider.
    when('/', {redirectTo: '/browse-actual'}).
    when('/browse-actual', {
      controller: 'BrowseActualCtrl as ctrl',
      templateUrl: 'browse-actual.html'
    }).
    when('/browse-snapshots', {
      controller: 'BrowseSnapshotsCtrl as ctrl',
      templateUrl: 'browse-snapshots.html' 
    }).
    when('/browse-snapshot-diff', {
      controller: 'BrowseSnapshotDiffCtrl as ctrl',
      templateUrl: 'browse-snapshot-diff.html'
    }).
    otherwise({redirectTo: '/browse-actual'});
}]);

zsd.config(['$httpProvider', function($httpProvider){
  $httpProvider.interceptors.push('HTTPErrorInterceptor');
  $httpProvider.interceptors.push('HTTPActivityInterceptor');
}]);

