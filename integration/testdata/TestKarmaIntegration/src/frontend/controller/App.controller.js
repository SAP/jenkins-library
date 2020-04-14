sap.ui.define([
   'sap/ui/core/mvc/Controller',
   'sap/m/MessageToast',
   'sap/ui/model/json/JSONModel'
], function (Controller, Toast, JSONModel) {
    'use strict';
    return Controller.extend('sap.ui.piper.controller.App', {

        onShowToast() {
            Toast.show('Opened Toast Message');
        },

        onShowDialog() {
            var oView = this.getView();
            var oDialog = oView.byId('demoDialog');
            // create dialog it doesn't already exist
            if (!oDialog) {
                // create dialog via fragment factory
                oDialog = sap.ui.xmlfragment(oView.getId(), 'sap.ui.piper.view.Dialog');
                oDialog.setBeginButton(new sap.m.Button({
                    text: 'Close',
                    press: function() {
                        oDialog.close();
                    }
                }));
                oView.addDependent(oDialog);
            }
            
            // load from nodejs server
            var model = new JSONModel();
            model.loadData('data');
            this.getView().setModel(model);

            oDialog.open();
        }
    });
});