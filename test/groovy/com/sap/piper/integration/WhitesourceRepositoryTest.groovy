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

        repository = new WhitesourceRepository(nullScript, [serviceUrl: "http://some.host.whitesource.com/api/"])
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(repository)
    }

    @After
    void tearDown() {
        printCallStack()
        nullScript.env = [:]
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

        repository.config['productName'] = "Correct Name Cloud"
        repository.config['projectNames'] = ["Correct Project Name", "Correct Project Name2"]

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

        repository.config['projectNames'] = ["Correct Project Name"]

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
        def config = [ whitesourceServiceUrl: "https://saas.whitesource.com/api", verbose: true]
        def requestBody = "{ \"someJson\" : { \"someObject\" : \"abcdef\" } }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.whitesourceServiceUrl,
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
        def config = [ serviceUrl: "https://saas.whitesource.com/api", verbose: true, userKey: "4711"]
        def requestBody = "{ \"someJson\" : { \"someObject\" : \"abcdef\" } }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.whitesourceServiceUrl,
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
        def config = [ whitesourceServiceUrl: "http://mo-323123123.sap.corp/some", verbose: false, userKey: "4711"]
        def requestBody = "{ \"someJson\" : { \"someObject\" : \"abcdef\" } }"

        def requestParams
        helper.registerAllowedMethod('httpRequest', [Map], { p ->
            requestParams = p
        })

        repository.httpWhitesource(requestBody)

        assertThat(requestParams, is(
            [
                url        : config.whitesourceServiceUrl,
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
            requestParams = p
            return [content: responseBody]
        })

        repository.fetchWhitesourceResource([httpMode: 'POST'])
        
    }

    @Test
    void testFetchReportForProduct() {
        repository.config.putAll([ whitesourceServiceUrl: "http://mo-323123123.sap.corp/some", verbose: false, productToken: "4711"])
        def command
        helper.registerAllowedMethod('sh', [String], { cmd ->
            command = cmd
        })

        repository.fetchReportForProduct("test.file")

        assertThat(command, equals('''#!/bin/sh -e
curl -o test.file -X POST http://some.host.whitesource.com/api/ -H 'Content-Type: application/json' -d '{
    "requestType": "getProductRiskReport",
    "productToken": "4711"
}'''
        ))
    }
}
