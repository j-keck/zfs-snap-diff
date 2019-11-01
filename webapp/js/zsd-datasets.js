angular.module('zsdDatasets', ['zsdServices']).
  directive('zsdDatasets', ['Config', function(Config){
    return {
      restrict: 'E',
      templateUrl: 'template-datasets.html',
      scope: {
        onDatasetSelected: '&',
        collapse: '='
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


        //
        // initializations
        //

        // auto-collapse if 'collapse' is defined
        if(angular.isDefined(scope.collapse)){
          scope.hideDatasets = true;
        }

        // select dataset if only one dataset is available
        if(scope.datasets.length == 1){
          scope.datasetSelected(scope.datasets[0]);
        }
      }
    }
}]);
