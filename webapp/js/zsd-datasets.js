angular.module('zsdDatasets', ['zsdServices']).
  directive('zsdDatasets', ['Config', function(Config){
    return {
      restrict: 'E',
      templateUrl: 'template-datasets.html',
      scope: {
        onDatasetSelected: '&'
      },
      link: function(scope, element, attrs){
        scope.datasets = Config.get('datasets');
        
        scope.datasetSelected = function(dataset){
          scope.hideDatasets = true;
          scope.onDatasetSelected({dataset: dataset});
        };

        scope.toggleHideDatasets = function(){
          scope.hideDatasets = ! scope.hideDatasets;
        };
      }
    }
}]);
