package com.sap.piper.integration

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.assertj.core.api.Assertions.assertThat
import static org.hamcrest.Matchers.is

class WhitesourceOrgAdminRepositoryTest extends BasePiperTest {

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jer)

    WhitesourceOrgAdminRepository repository

    @Before
    void init() throws Exception {
        repository = new WhitesourceOrgAdminRepository(nullScript, [serviceUrl: "http://some.host.whitesource.com/api/", verbose: true])
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(repository)
    }

    @After
    void tearDown() {
        printCallStack()
        nullScript.env = [:]
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

        repository.config['productName'] = "Correct Name Cloud"

        def result = repository.findProductMeta(whitesourceMetaResponse)

        assertThat(result).isEqualTo([
            token: '1111111-1111-1111-1111-111111111111',
            name : 'Correct Name Cloud'
        ])
    }

    @Test
    void testHttpWhitesourceInternalCallUserKey() {
        repository.orgAdminUserKey = "4711"
        def config = [ serviceUrl: "http://some.host.whitesource.com/api/", verbose: false, orgAdminUserKey: "4711"]
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
}
