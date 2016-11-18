angular
.module('appControllers')
.controller 'flowIndexCtrl', ['$scope', 'Notify', 'Flow', '$state', '$timeout'
($scope, Notify, Flow, $state, $timeout) ->
  vm = this

  tableCallback = {}
  vm.options =
    perPage: 20
    resource: Flow
    columns: [
      {
        name: '#'
        field: 'id'
      }
      {
        name: 'flow.name'
        field: 'name'
        clickCallback: ($event, item)->
          $event.preventDefault()
          $state.go('dashboard.flow.show', {id: item.id})
          false
      }
      {
        name: 'flow.created_at'
        field: 'created_at'
        template: '<span>{{item[field] | readableDateTime}}</span>'
      }
      {
        name: 'flow.update_at'
        field: 'update_at'
        template: '<span>{{item[field] | readableDateTime}}</span>'
      }
      {
        name: 'flow.status'
        width: '50px'
        template: "
<span class='label label-success' ng-if='item[\"status\"] == \"enabled\"'>{{'flow.enabled' | translate}}</span>
<span class='label label-default' ng-if='item[\"status\"] == \"disabled\"'>{{'flow.disabled' | translate}}</span>
"
        getStatus: (id)->
          $scope.flows[id]
      }
    ]
    menu:null
    callback: tableCallback
    onLoad: (result)->
      $timeout ()->
        $scope.getStatus().then (result)->
          $scope.flows = result.workflows
      , 500

  vm
]