package com.sap.piper.integration


import hudson.AbortException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.assertj.core.api.Assertions.assertThat
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isA

class WhitesourceRepositoryTest extends BasePiperTest {

    private ExpectedException exception = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(exception)
        .around(jlr)
        .around(jer)

    WhitesourceRepository repository

    @Before
    void init() throws Exception {
        nullScript.env['HTTP_PROXY'] = "http://proxy.wdf.sap.corp:8080"

        repository = new WhitesourceRepository(nullScript, [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/"]])
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(repository)
    }

    @After
    void tearDown() {
        printCallStack()
        nullScript.env = [:]
    }

    @Test
    void testMissingConfig() {
        exception.expect(AbortException)
        exception.expectMessage("Parameter 'whitesource.serviceUrl' must be provided as part of the configuration.")
        new WhitesourceRepository(nullScript, [:])
    }

    @Test
    void testResolveProjectsMeta() {
        def whitesourceMetaResponse = [
            projectVitals: [
                [
                    token: '410389ae-0269-4719-9cbf-fb5e299c8415',
                    name : 'NW'
                ],
                [
                    token: '2892f1db-4361-4e83-a89d-d28a262d65b9',
                    name : 'Correct Project Name2'
                ],
                [
                    token: '1111111-1111-1111-1111-111111111111',
                    name : 'Correct Project Name'
                ]
            ]
        ]

        repository.config.whitesource = [:]
        repository.config.whitesource['productName'] = "Correct Name Cloud"
        repository.config.whitesource['projectNames'] = ["Correct Project Name", "Correct Project Name2"]

        def result = repository.findProjectsMeta(whitesourceMetaResponse.projectVitals)

        assertThat(result, is(
            [
                {
                    token: '1111111-1111-1111-1111-111111111111'
                    name: 'Correct Name Cloud'
                },
                {
                    token: '2892f1db-4361-4e83-a89d-d28a262d65b9'
                    name: 'Correct Project Name2'
                }
            ]))

        assertThat(result.size(), 2)
    }

    @Test
    void testResolveProjectsMetaFailNotFound() {
        def whitesourceMetaResponse = [
            projectVitals: [
                [
                    token: '410389ae-0269-4719-9cbf-fb5e299c8415',
                    name : 'NW'
                ],
                [
                    token: '2892f1db-4361-4e83-a89d-d28a262d65b9',
                    name : 'Product Name'
                ],
                [
                    token: '1111111-1111-1111-1111-111111111111',
                    name : 'Product Name2'
                ]
            ]
        ]

        exception.expect(AbortException.class)

        exception.expectMessage("Correct Project Name")

        repository.config.putAll([whitesource : [projectNames: ["Correct Project Name"]]])

        repository.findProjectsMeta(whitesourceMetaResponse.projectVitals)
    }

    @Test
    void testSortLibrariesAlphabeticallyGAV() {

        def librariesResponse = [
            [
                groupId   : 'xyz',
                artifactId: 'abc'
            ],
            [
                groupId   : 'abc',
                artifactId: 'abc-def'
            ],
            [
                groupId   : 'abc',
                artifactId: 'def-abc'
            ],
            [
                groupId   : 'def',
                artifactId: 'test'
            ]
        ]

        repository.sortLibrariesAlphabeticallyGAV(librariesResponse)

        assertThat(librariesResponse, is(
            [
                {
                    groupId: 'abc'
                    artifactId: 'abc-def'
                },
                {
                    groupId: 'abc'
                    artifactId: 'def-abc'
                },
                {
                    groupId: 'def'
                    artifactId: 'test'
                },
                {
                    groupId: 'xyz'
                    artifactId: 'abc'
                }
            ]))
    }

    @Test
    void testSortVulnerabilitiesByScore() {

        def vulnerabilitiesResponse = [
            [
                vulnerability: [
                    score   : 6.9,
                    cvss3_score: 8.5
                ]
            ],
            [
                vulnerability: [
                    score   : 7.5,
                    cvss3_score: 9.8
                ]
            ],
            [
                vulnerability: [
                    score   : 4,
                    cvss3_score: 0
                ]
            ],
            [
                vulnerability: [
                    score   : 9.8,
                    cvss3_score: 0
                ]
            ],
            [
                vulnerability: [
                    score   : 0,
                    cvss3_score: 5
                ]
            ]
        ]

        repository.sortVulnerabilitiesByScore(vulnerabilitiesResponse)

        assertThat(vulnerabilitiesResponse, is(
            [
                {vulnerability: {
                    score: 9.8
                    cvss3_score: 0
                }}
,
                {vulnerability: {
                    score   : 7.5
                    cvss3_score: 9.8
                }}
,
                {vulnerability: {
                    score   : 6.9
                    cvss3_score: 8.5
                }}
,
                {vulnerability: {
                    score   : 0
                    cvss3_score: 5
                }}
,
                {vulnerability: {
                    score   : 4
                    cvss3_score: 0
                }}
            ]))
    }

    @Test
    void testHttpWhitesourceExternalCallNoUserKey() {
        def config = [whitesource: [serviceUrl: "https://saas.whitesource.com/api"], verbose: true]
        repository.config.putAll(config)
        def requestBody = "{ \"someJson\" : { \"someObject\" : \"abcdef\" } }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                proxy      : "http://proxy.wdf.sap.corp:8080"
            ]
        ))
    }

    @Test
    void testHttpWhitesourceExternalCallUserKey() {
        def config = [whitesource: [ serviceUrl: "https://saas.whitesource.com/api", userKey: "4711"], verbose: true]
        def requestBody = "{ \"someJson\" : { \"someObject\" : \"abcdef\" } }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                proxy      : "http://proxy.wdf.sap.corp:8080",
                userKey    : "4711"
            ]
        ))
    }

    @Test
    void testHttpWhitesourceInternalCallUserKey() {
        def config = [whitesource: [serviceUrl: "http://mo-323123123.sap.corp/some", userKey: "4711"], verbose: false]
        def requestBody = "{ \"someJson\" : { \"someObject\" : \"abcdef\" } }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : true
            ]
        ))
    }

    @Test
    void testHttpCallWithError() {
        def responseBody = """{
            \"errorCode\": 5001,
            \"errorMessage\": \"User is not allowed to perform this action\"
        }"""
        
        exception.expect(isA(AbortException.class))
        exception.expectMessage("[WhiteSource] Request failed with error message 'User is not allowed to perform this action' (5001)")
        
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            return [content: responseBody]
        })

        repository.fetchWhitesourceResource([httpMode: 'POST'])
        
    }

    @Test
    void testFetchReportForProduct() {
        repository.config.putAll([whitesource: [serviceUrl: "http://mo-323123123.sap.corp/some", productToken: "4712", userKey: "4711"], verbose: true])
        def requestBody = "{ \"requestType\": \"getProductRiskReport\", \"productToken\": \"${repository.config.whitesource.productToken}\" }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.fetchReportForProduct("test.file")

        assertThat(requestParams, is(
            [
                url           : repository.config.whitesource.serviceUrl,
                httpMode      : 'POST',
                acceptType    : 'APPLICATION_OCTETSTREAM',
                contentType   : 'APPLICATION_JSON',
                requestBody   : requestBody,
                quiet         : false,
                userKey       : repository.config.whitesource.userKey,
                httpProxy     : "http://test.sap.com:8080",
                outputFile    : "test.file",
                customHeaders : [[name: 'Cache-Control', value: 'no-cache, no-store, must-revalidate'], [name: 'Pragma', value: 'no-cache']]
            ]
        ))

        assertThat(jlr.log, containsString("Sending http request with parameters [requestType:getProductRiskReport, productToken:4711]"))
    }

    @Test
    void testFetchProductLicenseAlerts() {
        def config = [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/", userKey: "4711", productToken: "8547"]]
        nullScript.env['HTTP_PROXY'] = "http://test.sap.com:8080"
        repository.config.putAll(config)

        def requestBody = [
            requestType: "getProductAlertsByType",
            alertType: "REJECTED_BY_POLICY_RESOURCE",
            productToken: config.productToken
        ]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
            return [ content: "{ \"alerts\" : [] }"]
        })

        repository.fetchProductLicenseAlerts()

        assertThat(requestParams, is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                userKey    : config.whitesource.userKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))
    }

    @Test
    void testFetchProjectLicenseAlerts() {
        def projectToken = "8547"
        def config = [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/", userKey: "4711"]]
        repository.config.putAll(config)

        def requestBody = [
            requestType: "getProjectAlertsByType",
            alertType: "REJECTED_BY_POLICY_RESOURCE",
            projectToken: projectToken
        ]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
            return [ content: "{ \"alerts\" : [] }"]
        })

        repository.fetchProjectLicenseAlerts(projectToken)

        assertThat(requestParams, is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                userKey    : config.whitesource.userKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))
    }

    @Test
    void testFetchProjectsMetaInfo() {
        def config = [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/", userKey: "4711", productToken: '8475', projectNames: ['testProject1', 'testProject2']]]
        nullScript.env['HTTP_PROXY'] = "http://test.sap.com:8080"
        repository.config.putAll(config)

        def requestBody = [
            requestType: "getProductProjectVitals",
            productToken: config.productToken
        ]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
            return [ content: "{ \"projectVitals\" : [ { \"name\": \"testProject1\"}, { \"name\": \"testProject2\"} ] }"]
        })

        def result = repository.fetchProjectsMetaInfo()

        assertThat(requestParams, is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                userKey    : config.whitesource.userKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(result, is([[ name: "testProduct1"], [ name: "testProduct2"]]))
    }

    @Test
    void testFetchProjectsMetaInfoError() {
        def config = [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/", userKey: "4711", productName: 'kjdkjkhd', productToken: '8475', projectNames: ['testProject1', 'testProject2']]]
        repository.config.putAll(config)

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
            return [ content: "{ }"]
        })

        def errorCaught = false
        try {
            repository.fetchProjectsMetaInfo()
        } catch (e) {
            errorCaught = true
            assertThat(e, isA(AbortException.class))
            assertThat(e.getMessage(), is("[WhiteSource] Could not fetch any projects for product '${config.productName}' from backend, response was {}"))
        }
        assertThat(errorCaught, is(true))
    }


    @Test
    void testFetchVulnerabilitiesOnProjects() {
        def config = [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/", userKey: "4711", productToken: '8475', projectNames: ['testProject1', 'testProject2']]]
        nullScript.env['HTTP_PROXY'] = "http://test.sap.com:8080"
        repository.config.putAll(config)

        def requestBody1 = [
            requestType : "getProjectAlertsByType",
            alertType : "SECURITY_VULNERABILITY",
            projectToken: "1234"
        ]

        def requestBody2 = [
            requestType : "getProjectAlertsByType",
            alertType : "SECURITY_VULNERABILITY",
            projectToken: "2345"
        ]

        def requestParams = []
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams.add(p)
            return [ content: "{ \"alerts\" : [ { \"vulnerability\" : { \"cvss3_score\" : \"7\"} } ] }"]
        })

        def result = repository.fetchVulnerabilities([ [name: "testProject1", token: "1234"], [name: "testProject2", token: "2345"] ])

        assertThat(requestParams[0], is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody1,
                quiet      : false,
                userKey    : config.whitesource.userKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(requestParams[1], is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody2,
                quiet      : false,
                userKey    : config.whitesource.userKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(result.size(), is(2))
    }

    @Test
    void testFetchVulnerabilitiesOnProduct() {
        def config = [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/", userKey: "4711", productToken: '8475', productName : 'testProduct']]
        nullScript.env['HTTP_PROXY'] = "http://test.sap.com:8080"
        repository.config.putAll(config)

        def requestBody = [
            requestType : "getProductAlertsByType",
            alertType : "SECURITY_VULNERABILITY",
            productToken: config.productToken,
        ]

        def requestParams = []
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams.add(p)
            return [ content: "{ \"alerts\" : [ { \"vulnerability\" : { \"cvss3_score\" : \"7\"} } ] }"]
        })

        def result = repository.fetchVulnerabilities([ [name: "testProject1", token: "1234"], [name: "testProject2", token: "2345"] ])

        assertThat(requestParams[0], is(
            [
                url        : config.whitesource.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                userKey    : config.whitesource.userKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(result.size(), is(1))
    }
}
