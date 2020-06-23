sap.ui.define([
	'sap/ui/test/Opa5'
],
function (Opa5) {
	'use strict';


	function getFrameUrl(sHash, sUrlParameters) {
			sHash = sHash || "";
			var sUrl = jQuery.sap.getResourcePath("sap/ui/piper/index", ".html");
			if (sUrlParameters) {
				sUrlParameters = "?" + sUrlParameters;
			}
			return sUrl + sUrlParameters + "#" + sHash;
		}

	return Opa5.extend('sap.ui.piper.test.integration.pages.Common', {

		iStartMyApp : function (oOptions) {
			var sUrlParameters;
			oOptions = oOptions || { delay: 0 };

			sUrlParameters = "serverDelay=" + oOptions.delay;
			sUrlParameters += "&responderOn=true";
			sUrlParameters += "&sap-ui-language=en_US";

			this.iStartMyUIComponent({
				componentConfig: {
					name: "sap.ui.piper"
				}
			}
			);
		}
    });
});