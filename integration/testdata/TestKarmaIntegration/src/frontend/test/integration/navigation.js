/*global QUnit*/
/*global opaTest*/
 
sap.ui.require([
	'sap/ui/test/Opa5',
	'sap/ui/test/opaQunit',
	'sap/ui/piper/test/integration/pages/Common',
	'sap/ui/piper/test/integration/pages/App'
], function (Opa5, opaTest, Common) {
	'use strict';
 
	QUnit.module('Navigation', {
		beforeEach: function() {
			Opa5.extendConfig({
				arrangements: new Common(),
				viewNamespace: 'sap.ui.piper.view.',
				autoWait: true
			});
		}
	});
 
	opaTest('Should open the dialog', function (Given, When, Then) {
 
		// Arrangements
		Given.iStartMyApp();
 
		//Actions
		When.onTheAppPage.iPressTheOpenDialogButton();
 
		// Assertions
		Then.onTheAppPage.iShouldSeeADialog().
			and.iTeardownMyUIComponent();
	});

	opaTest('Should open a toast', function (Given, When, Then) {
 
		// Arrangements
		Given.iStartMyApp();
 
		//Actions
		When.onTheAppPage.iPressTheOpenToastButton();
 
		// Assertions
		Then.onTheAppPage.iShouldSeeAToast().
			and.iTeardownMyUIComponent();
	});
});