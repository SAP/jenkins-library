package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadFileRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class WhiteSourceConfigurationHelperTest extends BasePiperTest {
    JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/utilsTest/')
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jrfr)
        .around(jwfr)

    private static getMapping() {
        return [
            [name: "whitesourceOrgToken in ThisBuild", value: "config.orgToken", warnIfPresent: true],
            [name: "whitesourceProduct in ThisBuild", value: "config.whitesourceProductName"]
        ]
    }

    @Before
    void init() {
        helper.registerAllowedMethod('readJSON', [Map], {return [:]})
        helper.registerAllowedMethod('readProperties', [Map], {return new Properties()})
    }

    @Test
    void testReadScalaConfig() {
        def resMap = WhitesourceConfigurationHelper.readScalaConfig(nullScript, getMapping(), "build.sbt")
        assertThat(resMap, hasKey(WhitesourceConfigurationHelper.SCALA_CONTENT_KEY))

        // mapping tokens should be removed from parsed content
        assertThat(resMap[WhitesourceConfigurationHelper.SCALA_CONTENT_KEY], not(hasItem(containsString("whitesourceOrgToken in ThisBuild"))))

        assertThat(resMap, hasEntry("whitesourceOrgToken in ThisBuild", "\"org-token\""))
        assertThat(resMap, hasEntry("whitesourceProduct in ThisBuild", "\"PRODUCT VERSION\""))
    }

    @Test
    void testSerializeScalaConfig() {
        def resMap = [
            "whitesourceOrgToken in ThisBuild": "\"some-long-hash-value\"",
            "whitesourceProduct in ThisBuild": "\"PRODUCT IDENTIFIER\"",
            "whitesourceServiceUrl in ThisBuild": "uri(\"http://mo-393ef744d.mo.sap.corp:8080/wsui/wspluginProxy.jsp\")"
        ]
        resMap[WhitesourceConfigurationHelper.SCALA_CONTENT_KEY] = ["// build.sbt -- scala build file", "name := \"minimal-scala\"", "libraryDependencies += \"org.scalatest\" %% \"scalatest\" % \"2.2.4\" % \"test\""]
        def fileContent = WhitesourceConfigurationHelper.serializeScalaConfig(resMap)

        resMap[WhitesourceConfigurationHelper.SCALA_CONTENT_KEY].each {
            line ->
                assertThat(fileContent, containsString("${line}\r"))
        }

        assertThat(fileContent, containsString("whitesourceOrgToken in ThisBuild := \"some-long-hash-value\""))
        assertThat(fileContent, containsString("whitesourceProduct in ThisBuild := \"PRODUCT IDENTIFIER\""))
        assertThat(fileContent, containsString("whitesourceServiceUrl in ThisBuild := uri(\"http://mo-393ef744d.mo.sap.corp:8080/wsui/wspluginProxy.jsp\")"))
    }

    @Test
    void testExtendConfigurationFileUnifiedAgent() {
        WhitesourceConfigurationHelper.extendConfigurationFile(nullScript, utils, [scanType: 'unifiedAgent', configFilePath: './config', orgToken: 'abcd', productName: 'name', productToken: '1234', userKey: '0000'], "./")
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("apiKey=abcd"))
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("productName=name"))
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("productToken=1234"))
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("userKey=0000"))
    }

    @Test
    void testExtendConfigurationFileNpm() {
        WhitesourceConfigurationHelper.extendConfigurationFile(nullScript, utils, [scanType: 'npm', configFilePath: './config', orgToken: 'abcd', productName: 'name', productToken: '1234', userKey: '0000'], "./")
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("\"apiKey\": \"abcd\","))
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("\"productName\": \"name\","))
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("\"productToken\": \"1234\","))
        assertThat(jwfr.files['./config.c92a71303bcc841344e07d1bf49d1f9b'], containsString("\"userKey\": \"0000\""))
    }

    @Test
    void testExtendConfigurationFilePip() {
        WhitesourceConfigurationHelper.extendConfigurationFile(nullScript, utils, [scanType: 'pip', configFilePath: './setup.py', orgToken: 'abcd', productName: 'name', productToken: '1234', userKey: '0000'], "./")
        assertThat(jwfr.files['./setup.py.8813e60e0d9f7cacf0c414ae4964816f.py'], containsString("'org_token': 'abcd',"))
        assertThat(jwfr.files['./setup.py.8813e60e0d9f7cacf0c414ae4964816f.py'], containsString("'product_name': 'name',"))
        assertThat(jwfr.files['./setup.py.8813e60e0d9f7cacf0c414ae4964816f.py'], containsString("'product_token': '1234',"))
        assertThat(jwfr.files['./setup.py.8813e60e0d9f7cacf0c414ae4964816f.py'], containsString("'user_key': '0000'"))
    }

    @Test
    void testExtendConfigurationFileSbt() {
        WhitesourceConfigurationHelper.extendConfigurationFile(nullScript, utils, [scanType: 'sbt', configFilePath: './build.sbt', orgToken: 'abcd', productName: 'name', productToken: '1234', userKey: '0000', agentUrl: 'http://mo-393ef744d.mo.sap.corp:8080/wsui/wspluginProxy.jsp'], "./")
        assertThat(jwfr.files['./build.sbt'], containsString("whitesourceOrgToken in ThisBuild := \"abcd\""))
        assertThat(jwfr.files['./build.sbt'], containsString("whitesourceProduct in ThisBuild := \"name\""))
        assertThat(jwfr.files['./build.sbt'], containsString("whitesourceServiceUrl in ThisBuild := uri(\"http://mo-393ef744d.mo.sap.corp:8080/wsui/wspluginProxy.jsp\")"))
    }
}

