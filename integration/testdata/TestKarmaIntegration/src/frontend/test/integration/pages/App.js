sap.ui.require([
	'sap/ui/test/Opa5'
],
function (Opa5) {
	'use strict';

	Opa5.createPageObjects({

		onTheAppPage: {
			actions: {
				iPressTheOpenDialogButton: function () {
					return this.waitFor({
						controlType: 'sap.m.Button',
						success: function (aButtons) {
							aButtons[1].$().trigger('tap');
						},
						errorMessage: 'Did not find the showDialogButton button on the app page'
					});
				},
                iPressTheOpenToastButton: function () {
					return this.waitFor({
						controlType: 'sap.m.Button',
						success: function (aButtons) {
							aButtons[0].$().trigger('tap');
						},
						errorMessage: 'Did not find the showToastButton button on the app page'
					});
				}
			},
			assertions: {
				iShouldSeeADialog: function () {
					return this.waitFor({
						controlType: 'sap.m.Dialog',
						success: function () {
							Opa5.assert.ok(true, 'The dialog is open');
						},
						errorMessage: 'Did not find the dialog control'
					});
				},
                iShouldSeeAToast: function () {
					return this.waitFor({
						pollingInterval : 100,
						matchers: function () {
							return  jQuery(".sapMMessageToast").text();
						},
						success : function () {
							ok(true, 'Found a Toast');
							//Opa5.assert.ok(true, 'The dialog is open');
						},
						errorMessage : 'No Toast message detected!'
					});
				}
			}
		}
	});
});