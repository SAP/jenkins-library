// karma.conf.js
module.exports = function(config) {
  config.set({
    basePath : "./",
        frameworks : [ "qunit", "openui5" ],
        files : [       
            {
                pattern : "src/frontend/test/karma-qunit.js", included: true
            },
            {
                pattern : "src/frontend/test/integration/AllJourneys.js", included: true
            },
            {
                pattern : "src/frontend/**/*", included : false
            }
        ],
        browsers: ['chrome_selenium'],

        hostname: 'karma',
        customLaunchers: {
            chrome_selenium: {
                base: 'WebDriver',
                config: {
                     hostname: 'selenium',
                     port: 4444
                },
                browserName: 'chrome',
                name: 'Chrome'
            },
        },
        port : 9876,
        logLevel : "DEBUG",
        autoWatch : false,
        singleRun : true,
        browserNoActivityTimeout : 40000,
        browserDisconnectTolerance: 2,

        openui5 : {
            path : "https://sapui5.hana.ondemand.com/resources/sap-ui-core.js"
        },
        client: {
            openui5 : {
                config : {
                    theme : "sap_bluecrystal",
                    resourceroots : {
                    'sap.ui.piper.test': '/base/src/frontend/test',
                    'sap.ui.piper.controller': '/base/src/frontend/controller',
                    'sap.ui.piper': '/base/src/frontend'
                    }
                }
            },
        }
    }   
  );
};