package com.sap.piper.integration

import hudson.AbortException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsErrorRule
import util.JenkinsLoggingRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.assertj.core.api.Assertions.assertThat
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isA

class WhitesourceOrgAdminRepositoryTest extends BasePiperTest {

    private JenkinsErrorRule thrown = new JenkinsErrorRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(loggingRule)
        .around(environmentRule)

    WhitesourceOrgAdminRepository repository

    @Before
    void init() throws Exception {
        repository = new WhitesourceOrgAdminRepository(nullScript, [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/"], verbose: true])
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(repository)
    }

    @After
    void tearDown() {
        printCallStack()
        nullScript.env = [:]
    }

    @Test
    void testMissingConfig() {
        def errorCaught = false
        try {
            new WhitesourceOrgAdminRepository(nullScript, [:])
        } catch (e) {
            errorCaught = true
            assertThat(e, isA(AbortException.class))
            assertThat(e.getMessage(), is("Parameter 'whitesource.serviceUrl' must be provided as part of the configuration."))
        }
        assertThat(errorCaught, is(true))
    }

    @Test
    void testAccessor() {
        new WhitesourceOrgAdminRepository(nullScript, [whitesourceAccessor: "com.sap.piper.integration.WhitesourceRepository", whitesource: [serviceUrl: "http://test.com"]])
    }

    @Test
    void testResolveProductMeta() {

        def whitesourceMetaResponse = [
            productVitals: [
                [
                    token: '410389ae-0269-4719-9cbf-fb5e299c8415',
                    name : 'NW'
                ],
                [
                    token: '2892f1db-4361-4e83-a89d-d28a262d65b9',
                    name : 'XS UAA'
                ],
                [
                    token: '1111111-1111-1111-1111-111111111111',
                    name : 'Correct Name Cloud'
                ]
            ]
        ]

        repository.config.putAll([whitesource: [productName: "Correct Name Cloud"]])

        def result = repository.findProductMeta(whitesourceMetaResponse)

        assertThat(result).isEqualTo([
            token: '1111111-1111-1111-1111-111111111111',
            name : 'Correct Name Cloud'
        ])
    }

    @Test
    void testHttpWhitesourceInternalCallUserKey() {
        def config = [whitesource: [ serviceUrl: "http://some.host.whitesource.com/api/", orgAdminUserKey: "4711"], verbose: false]
        repository.config.putAll(config)
        def requestBody = ["someJson" : [ "someObject" : "abcdef" ]]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : true,
                userKey    : config.orgAdminUserKey
            ]
        ))
    }

    @Test
    void testHttpWhitesourceInternalCallUserKeyVerboseProxy() {
        def config = [whitesource: [ serviceUrl: "http://some.host.whitesource.com/api/", orgAdminUserKey: "4711"], verbose: true]
        nullScript.env['HTTP_PROXY'] = "http://test.sap.com:8080"
        repository.config.putAll(config)
        def requestBody = ["someJson" : [ "someObject" : "abcdef" ]]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                userKey    : config.orgAdminUserKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(loggingRule.log, containsString("Sending http request with parameters"))
        assertThat(loggingRule.log, containsString("Received response"))
    }

    @Test
    void testCreateProduct() {
        def config = [
            whitesource: [
                    serviceUrl: "http://some.host.whitesource.com/api/",
                    verbose: false,
                    orgAdminUserKey: "4711",
                    orgToken: "abcd1234",
                    productName: "testProduct",
                    emailAddressesOfInitialProductAdmins: ['some@somewhere.com', 'some2@somewhere.com']
                ]
        ]
        repository.config.putAll(config)
        def requestBody1 = [
            requestType: "getOrganizationProductVitals",
            orgToken: config.orgToken,
            userKey: "4711"
        ]

        def requestBody2 = [
            "requestType" : "setProductAssignments",
            "productToken" : "54785",
            "productMembership" : ["userAssignments":[], "groupAssignments":[]],
            "productAdmins" : ["userAssignments":[[ "email": "some@somewhere.com" ], ["email": "some2@somewhere.com"]]],
            "alertsEmailReceivers" : ["userAssignments":[]],
            "userKey": "4711"
        ]

        def requestParams = []
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams.add(p)
            return [ content : "{ \"productToken\" : \"54785\" }" ]
        })

        repository.createProduct()

        assertThat(requestParams[0], is(
            [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody1,
                quiet      : false,
                userKey    : config.orgAdminUserKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(requestParams[1], is(
            [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody2,
                quiet      : false,
                userKey    : config.orgAdminUserKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))
    }

    @Test
    void testIssueHttpRequestError() {
        def config = [whitesource: [ serviceUrl: "http://some.host.whitesource.com/api/", orgAdminUserKey: "4711"], verbose: false]
        repository.config.putAll(config)
        def requestBody = ["someJson" : [ "someObject" : "abcdef" ]]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
            return [content: "{ \"errorCode\" : \"4546\", \"errorMessage\" : \"some text\" } }"]
        })

        def errorCaught = false
        try {
            repository.issueHttpRequest(requestBody)
        } catch (e) {
            errorCaught = true
            assertThat(e, isA(AbortException.class))
            assertThat(e.getMessage(), equals("[WhiteSource] Request failed with error message 'some text' (4546)."))
        }
        assertThat(errorCaught, is(true))

        assertThat(requestParams, is(
            [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : true,
                userKey    : config.orgAdminUserKey
            ]
        ))
    }

    @Test
    void testFetchProductMetaInfo() {
        def config = [whitesource: [ serviceUrl: "http://some.host.whitesource.com/api/", orgAdminUserKey: "4711", orgToken: "12345", productName: "testProduct"], verbose: true]
        nullScript.env['HTTP_PROXY'] = "http://test.sap.com:8080"
        repository.config.putAll(config)

        def requestBody = [
            requestType: "getOrganizationProductVitals",
            orgToken: config.orgToken,
            userKey: "4711"
        ]

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
            return [ content: "{ \"productVitals\" : [ { \"name\": \"testProduct\"} ] }"]
        })

        def result = repository.fetchProductMetaInfo()

        assertThat(requestParams, is(
            [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: requestBody,
                quiet      : false,
                userKey    : config.orgAdminUserKey,
                httpProxy  : "http://test.sap.com:8080"
            ]
        ))

        assertThat(result, is([ name: "testProduct"]))
        assertThat(loggingRule.log, containsString("Sending http request with parameters"))
        assertThat(loggingRule.log, containsString("Received response"))
    }
}
