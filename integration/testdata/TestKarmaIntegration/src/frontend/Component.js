sap.ui.define([
   'sap/ui/core/UIComponent',
   'sap/ui/model/json/JSONModel',
   'sap/ui/model/resource/ResourceModel'
], function (UIComponent, JSONModel, ResourceModel) {
   'use strict';
   return UIComponent.extend('sap.ui.piper.Component', {
            metadata : {
		rootView: 'sap.ui.piper.view.App'
	},
      init : function () {
         // call the init function of the parent
         UIComponent.prototype.init.apply(this, arguments);
      }
   });
});