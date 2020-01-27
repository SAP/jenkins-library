#!groovy
package steps

import com.sap.piper.internal.integration.Protecode
import groovy.json.JsonSlurper
import hudson.AbortException
import org.junit.Assert
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.springframework.beans.factory.annotation.Autowired
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.containsString

import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.isA
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.not
import static org.hamcrest.CoreMatchers.nullValue
import static org.hamcrest.CoreMatchers.anyOf

import static org.junit.Assert.assertThat

class protecodeExecuteScanTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    public JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    public JenkinsStepRule jsr = new JenkinsStepRule(this)
    public JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    def dockerMockArgs = [:]
    class DockerMock {
        DockerMock(name){
            dockerMockArgs.name = name
        }
        def withRegistry(paramRegistry, paramCredentials, paramClosure){
            dockerMockArgs.paramRegistry = paramRegistry
            dockerMockArgs.paramCredentials = paramCredentials
            return paramClosure()
        }
        def withRegistry(paramRegistry, paramClosure){
            dockerMockArgs.paramRegistryAnonymous = paramRegistry.toString()
            return paramClosure()
        }

        def image(name) {
            dockerMockArgs.name = name
            return new DockerImageMock()
        }
    }

    def dockerMockPushes = []
    def dockerMockPull = false
    class DockerImageMock {
        DockerImageMock(){}
        def push(tag){
            dockerMockPushes.add(tag)
        }
        def push(){
            push('default')
        }

        def pull(){
            dockerMockPull = true
        }
    }

    def httpRequestCount = 0
    def httpRequestTrace  = [:]

    def invocations = 0

    @Autowired
    Protecode protecodeStub

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(exception)
        .around(jlr)
        .around(jscr)
        .around(jsr)
        .around(jer)

    @Before
    void init() {
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            Assert.assertEquals('protecodeCreds', l[0].credentialsId)
            getBinding().setProperty('user', 'test_user')
            getBinding().setProperty('password', '**********')
            try {
                c()
            } finally {
                getBinding().setProperty('user', null)
                getBinding().setProperty('password', null)
            }
        })

        helper.registerAllowedMethod('removeJobSideBarLinks', [String], {
            url ->
                Assert.assertEquals("artifact/protecode_report.pdf", url.toString())
        })

        helper.registerAllowedMethod('timeout', [Integer, Closure], {
            numMinutes, closure ->
                Assert.assertEquals(60, numMinutes)
                closure()
        })

        helper.registerAllowedMethod('addJobSideBarLink', [String, String, String], {
            url, name, icon ->
                Assert.assertEquals("artifact/protecode_report.pdf", url.toString())
                Assert.assertEquals("Protecode Report", name)
                Assert.assertEquals("images/24x24/graph.png", icon)
        })

        helper.registerAllowedMethod('addRunSideBarLink', [String, String, String], {
            url, name, icon ->
                Assert.assertThat(url.toString(), anyOf(
                    is('artifact/protecode_report.pdf'),
                    is('https://protecode.c.eu-de-2.cloud.sap/products/4711/'))
                )
                Assert.assertThat(name, anyOf(
                    is('Protecode Report'),
                    is('Protecode WebUI'))
                )
                Assert.assertEquals("images/24x24/graph.png", icon)
        })

        helper.registerAllowedMethod('archiveArtifacts', [Map], {
            m ->
                Assert.assertEquals("protecode_report.pdf", m.artifacts.toString())
                Assert.assertEquals(false, m["allowEmptyArchive"])
        })

        invocations = 0
        helper.registerAllowedMethod('readJSON', [Map], {
            m ->
                invocations++
                if(invocations == 1)
                    return [results: [product_id: '4711', report_url: 'https://protecode.c.eu-de-2.cloud.sap/api/products/4711']]
                if(invocations == 2) {
                    return new JsonSlurper().parse(new File("test/resources/executeProtecodeScanTest/protecode_result_no_violations.json"))
                }
        })

        httpRequestCount = 0
        helper.registerAllowedMethod('httpRequest', [Map], {
            m ->
                httpRequestTrace.put(httpRequestCount, m)
                httpRequestCount++
                return new Object(){
                    def getStatus() {
                        return 200
                    }
                    def getContent() {
                        return "HTTP response body"
                    }
                }
        })

        helper.registerAllowedMethod('writeFile', [Map], null)

        jscr.setReturnValue("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")

        binding.setVariable('docker', new DockerMock('test'))
        nullScript.binding.setVariable('docker', new DockerMock('test'))

        nullScript.binding.setVariable('env', [ON_K8S: 'false'])
    }


    @Test
    void testDefaultValuesNoFindings() {
        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void testOverwriteFilePath() {
        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            dockerImage: 'path/myTestImage:tag',
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
    }

    @Test
    void testCustomImageNoFindings() {
        jscr.setReturnValue("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_myTestImage.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            dockerImage: 'path/myTestImage:tag',
            dockerRegistryUrl: 'https://testRegistry:55555',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://testRegistry:55555'))
        assertThat(jscr.shell, hasItem('docker pull testRegistry:55555/path/myTestImage:tag && docker save --output path_myTestImage.tar testRegistry:55555/path/myTestImage:tag'))
        assertThat(jscr.shell, hasItem("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_myTestImage.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
    }

    @Test
    void testCustomImageNoFindingsKubernetes() {
        jscr.setReturnValue("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_myTestImage.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        Map dockerKuberMap
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], {m, body ->
            dockerKuberMap = m
            return body()
        })

        helper.registerAllowedMethod('container', [String.class, Closure.class], {s, body ->
            assertThat(s, is('skopeo'))
            return body()
        })

        nullScript.binding.setVariable('env', [ON_K8S: 'true'])

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            dockerImage: 'path/myTestImage:tag',
            dockerRegistryUrl: 'https://testRegistry:55555',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        assertThat(dockerMockArgs.paramRegistryAnonymous, nullValue())
        assertThat(jscr.shell, not(hasItem('docker pull testRegistry:55555/path/myTestImage:tag && docker save --output path_myTestImage.tar testRegistry:55555/path/myTestImage:tag')))

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false docker://testRegistry:55555/path/myTestImage:tag docker-archive:path_myTestImage.tar:path/myTestImage:tag'))

        assertThat(jscr.shell, hasItem("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_myTestImage.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
    }

    @Test
    void testFetchUrl() {
        def url = 'http://get.videolan.org/vlc/2.2.1/macosx/vlc-2.2.1.dmg'
        jscr.setReturnValue("#!/bin/sh -e curl -X POST -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -H 'Url: ${url}' https://protecode.c.eu-de-2.cloud.sap/api/fetch/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            fetchUrl: url,
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
        assertThat(jscr.shell, hasItem("#!/bin/sh -e curl -X POST -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -H 'Url: ${url}' https://protecode.c.eu-de-2.cloud.sap/api/fetch/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
    }

    @Test
    void testCustomImageWithDockerMetadataNoFindings() {

        nullScript.globalPipelineEnvironment.setDockerMetadata([
            repo: 'testRegistry:55555',
            tag_name: 'testRegistry:55555/path/testImage:tag',
            image_name: 'testRegistry:55555/path/testImage:tag'
        ])

        jscr.setReturnValue("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_testImage.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://testRegistry:55555'))
        assertThat(jscr.shell, hasItem('docker pull testRegistry:55555/path/testImage:tag && docker save --output path_testImage.tar testRegistry:55555/path/testImage:tag'))
        assertThat(jscr.shell, hasItem("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_testImage.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
    }

    @Test
    void testCustomImageWithAppContainerDockerMetadata() {

        nullScript.globalPipelineEnvironment.setAppContainerDockerMetadata([
            repo: 'testRegistryX:55555',
            tag_name: 'testRegistryX:55555/path/testImageX:tag',
            image_name: 'testRegistryX:55555/path/testImageX:tag'
        ])

        jscr.setReturnValue("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_testImageX.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://testRegistryX:55555'))
        assertThat(jscr.shell, hasItem('docker pull testRegistryX:55555/path/testImageX:tag && docker save --output path_testImageX.tar testRegistryX:55555/path/testImageX:tag'))
        assertThat(jscr.shell, hasItem("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T path_testImageX.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
    }

    @Test
    void testOtherValuesWithFindings() {

        def invocations = 0
        helper.registerAllowedMethod('readJSON', [Map], {
            m ->
                invocations++
                if(invocations == 1)
                    return [results: [product_id: '4711', report_url: 'https://protecode.c.eu-de-2.cloud.sap/api/products/4711']]
                if(invocations == 2)
                    return new JsonSlurper().parse(new File("test/resources/executeProtecodeScanTest/protecode_result_violations.json"))
        })

        exception.expect(isA(AbortException.class))
        exception.expectMessage(containsString("Protecode detected Open Source Software Security vulnerabilities, the project is not compliant. For details see the archived report or the web ui: https://protecode.c.eu-de-2.cloud.sap/products/4711/"))

        try{
            jsr.step.executeProtecodeScan([
                script: nullScript,
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                protecodeStub: protecodeStub,
                filePath: 'test_image.tar',
                protecodeCredentialsId: 'protecodeCreds',
                protecodeExcludeCVEs: ['CVE-2018-1', 'CVE-2017-1'],
                protecodeGroup: '16'
            ])
        }finally{
            assertThat(jlr.log, containsString(
                "227 Known vulnerabilities were found during the scan! of which 13 had a CVSS v2 score >= 7.0 and 129 had a CVSS v3 score >= 7.0.\n" +
                "2 vulnerabilities were excluded via configuration ([CVE-2018-1, CVE-2017-1]) and 0 vulnerabilities were triaged via the webUI.\n" +
                "In addition 1125 historical vulnerabilities were spotted."
            ))
        }
    }

    @Test
    void testTriaging() {

        def invocations = 0
        helper.registerAllowedMethod('readJSON', [Map], {
            m ->
                invocations++
                if(invocations == 1)
                    return [results: [product_id: '4711', report_url: 'https://protecode.c.eu-de-2.cloud.sap/api/products/4711']]
                if(invocations == 2)
                    return new JsonSlurper().parse(new File("test/resources/executeProtecodeScanTest/protecode_result_triaging.json"))
        })

        exception.expect(isA(AbortException.class))
        exception.expectMessage(containsString("Protecode detected Open Source Software Security vulnerabilities, the project is not compliant. For details see the archived report or the web ui: https://protecode.c.eu-de-2.cloud.sap/products/4711/"))

        try{
            jsr.step.executeProtecodeScan([
                script: nullScript,
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                protecodeStub: protecodeStub,
                filePath: 'test_image.tar',
                protecodeCredentialsId: 'protecodeCreds',
                protecodeGroup: '16'
            ])
        }finally{
            assertThat(jlr.log, containsString(
                "36 Known vulnerabilities were found during the scan! of which 0 had a CVSS v2 score >= 7.0 and 15 had a CVSS v3 score >= 7.0.\n" +
                "0 vulnerabilities were excluded via configuration ([]) and 187 vulnerabilities were triaged via the webUI.\n" +
                "In addition 1132 historical vulnerabilities were spotted."
            ))
        }
    }

    @Test
    void testWithReusingExistingReport() {

        def invocations = 0
        helper.registerAllowedMethod('readJSON', [Map], {
            m ->
                invocations++
                if(invocations == 1)
                    return [meta: [code: 200], products: [[id: 4711, status: 'R', sha1sum: 'c6e717ae06123ee75d4faffd7baf48fa48a63170', product_id: 4711, name: 'test_image.tar', custom_data: []]]]
                if(invocations == 2) {
                    return new JsonSlurper().parse(new File("test/resources/executeProtecodeScanTest/protecode_result_no_violations.json"))
                }
        })

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16',
            reuseExisting: true
        ])

        Assert.assertFalse(jscr.shell.contains("#!/bin/sh -e curl -H 'Authorization :Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/apps/16/?q=file:test_image.tar"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[2].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void testWithAttemptedReuse() {

        def invocations = 0
        helper.registerAllowedMethod('readJSON', [Map], {
            m ->
                invocations++
                if(invocations == 1)
                    return [meta: [code: 200], products: []]
                if(invocations == 2)
                    return [results: [product_id: '4711', report_url: 'https://protecode.c.eu-de-2.cloud.sap/api/products/4711']]
                if(invocations == 3) {
                    return new JsonSlurper().parse(new File("test/resources/executeProtecodeScanTest/protecode_result_no_violations.json"))
                }
        })

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            reuseExisting: true,
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])
        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/apps/16/?q=file:test_image.tar"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[2].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void disableSidebarLinks() {

        helper.registerAllowedMethod('removeJobSideBarLinks', [String], {
            url ->
                Assert.fail("Should not be called")
        })

        helper.registerAllowedMethod('addJobSideBarLink', [String, String, String], {
            url, name, icon ->
                Assert.fail("Should not be called")
        })

        helper.registerAllowedMethod('addRunSideBarLink', [String, String, String], {
            url, name, icon ->
                Assert.fail("Should not be called")
        })

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            addSideBarLink: false,
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        Assert.assertThat(jscr.shell, hasItem("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void renameReportFile() {

        helper.registerAllowedMethod('removeJobSideBarLinks', [String], {
            url ->
                Assert.assertEquals("artifact/test_image_report.pdf", url.toString())
        })

        helper.registerAllowedMethod('addJobSideBarLink', [String, String, String], {
            url, name, icon ->
                Assert.assertEquals("artifact/test_image_report.pdf", url.toString())
                Assert.assertEquals("Protecode Report", name)
                Assert.assertEquals("images/24x24/graph.png", icon)
        })

        helper.registerAllowedMethod('addRunSideBarLink', [String, String, String], {
            url, name, icon ->
                Assert.assertThat(url.toString(), anyOf(
                    is('artifact/test_image_report.pdf'),
                    is('https://protecode.c.eu-de-2.cloud.sap/products/4711/'))
                )
                Assert.assertThat(name, anyOf(
                    is('Protecode Report'),
                    is('Protecode WebUI'))
                )
                Assert.assertEquals("images/24x24/graph.png", icon)
        })

        helper.registerAllowedMethod('archiveArtifacts', [Map], {
            m ->
                Assert.assertEquals("test_image_report.pdf", m.artifacts.toString())
                Assert.assertEquals(false, m["allowEmptyArchive"])
        })

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16',
            reportFileName: 'test_image_report.pdf'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void testCompleteCleanupMode() {

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16',
            cleanupMode: 'complete'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertEquals(httpRequestTrace.size(), 3)
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
        Assert.assertThat(httpRequestTrace[2].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[2].httpMode.toString(), is("DELETE"))
    }

    @Test
    void testBinaryCleanupMode() {

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16',
            cleanupMode: 'binary'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertEquals(httpRequestTrace.size(), 2)
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void testDefaultCleanupMode() {

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: true' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertEquals(httpRequestTrace.size(), 2)
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

    @Test
    void testNoneCleanupMode() {

        jscr.setReturnValue("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: false' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'", "${Protecode.DELIMITER}status=200")

        jsr.step.executeProtecodeScan([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            protecodeStub: protecodeStub,
            filePath: 'test_image.tar',
            protecodeCredentialsId: 'protecodeCreds',
            protecodeGroup: '16',
            cleanupMode: 'none'
        ])

        Assert.assertTrue(jscr.shell.contains("#!/bin/sh -e curl --insecure -H 'Authorization: Basic dGVzdF91c2VyOioqKioqKioqKio=' -H 'Group: 16' -H 'Delete-Binary: false' -T test_image.tar https://protecode.c.eu-de-2.cloud.sap/api/upload/ --write-out '${Protecode.DELIMITER}status=%{http_code}'".toString()))
        Assert.assertEquals(httpRequestTrace.size(), 2)
        Assert.assertThat(httpRequestTrace[0].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/"))
        Assert.assertThat(httpRequestTrace[1].url.toString(), is("https://protecode.c.eu-de-2.cloud.sap/api/product/4711/pdf-report"))
    }

}
