#!groovy
import com.sap.piper.DescriptorUtils
import com.sap.piper.JsonUtils
import com.sap.piper.integration.WhitesourceOrgAdminRepository
import com.sap.piper.integration.WhitesourceRepository
import hudson.AbortException
import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.springframework.beans.factory.annotation.Autowired
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class WhitesourceExecuteScanTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(loggingRule)
        .around(writeFileRule)
        .around(stepRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('ID-123456789', 'token-0815')
            .withCredentials('ID-9876543', 'token-0816'))

    def whitesourceOrgAdminRepositoryStub
    def whitesourceStub

    @Autowired
    DescriptorUtils descriptorUtilsStub

    @Before
    void init() {
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], { m ->
            if (m.artifacts == null) {
                throw new Exception('artifacts cannot be null')
            }
            return null
        })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            return [].toArray()
        })

        whitesourceOrgAdminRepositoryStub = new WhitesourceOrgAdminRepository(nullScript, [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/"]])
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(whitesourceOrgAdminRepositoryStub)

        whitesourceStub = new WhitesourceRepository(nullScript, [whitesource: [serviceUrl: "http://some.host.whitesource.com/api/"]])
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(whitesourceStub)

        helper.registerAllowedMethod("fetchProductMetaInfo", [], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"productVitals\": [{ \"id\": 59639, \"name\": \"SHC - Piper\", \"token\": \"e30132d8e8f04a4c8be6332c75a0ff0580ab326fa7534540ad326e97a74d945b\", \"creationDate\": \"2017-09-20 09:22:46 +0000\", \"lastUpdatedDate\": \"2018-09-19 09:44:40 +0000\" }]}")
        })
        helper.registerAllowedMethod("fetchProjectsMetaInfo", [], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"projectVitals\": [{ \"id\": 261964, \"name\": \"piper-demo - 0.0.1\", \"token\": \"a2a62e5d7beb4170ad4dccfa3316b5a4cd3fadefc56c49f88fbf9400a09f7d94\", \"creationDate\": \"2017-09-21 00:28:06 +0000\", \"lastUpdatedDate\": \"2017-10-12 01:03:05 +0000\" }]}").projectVitals
        })
        helper.registerAllowedMethod("fetchReportForProduct", [String], { })
        helper.registerAllowedMethod( "fetchProjectLicenseAlerts", [Object.class], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [] }").alerts
        })
        helper.registerAllowedMethod( "fetchProductLicenseAlerts", [], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [] }").alerts
        })
        helper.registerAllowedMethod( "fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [] }").alerts
        })
        helper.registerAllowedMethod( "createProduct", [], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"productToken\": \"e30132d8e8f04a4c8be6332c75a0ff0580ab326fa7534540ad326e97a74d945b\" }")
        })
        helper.registerAllowedMethod( "publishHTML", [Map], {})

        helper.registerAllowedMethod( "getNpmGAV", [String], {return [group: 'com.sap.node', artifact: 'test-node', version: '1.2.3']})
        helper.registerAllowedMethod( "getSbtGAV", [String], {return [group: 'com.sap.sbt', artifact: 'test-scala', version: '1.2.3']})
        helper.registerAllowedMethod( "getPipGAV", [String], {return [artifact: 'test-python', version: '1.2.3']})
        helper.registerAllowedMethod( "getMavenGAV", [String], {return [group: 'com.sap.maven', artifact: 'test-java', version: '1.2.3']})

        nullScript.commonPipelineEnvironment.configuration = nullScript.commonPipelineEnvironment.configuration ?: [:]
        nullScript.commonPipelineEnvironment.configuration['steps'] = nullScript.commonPipelineEnvironment.configuration['steps'] ?: [:]
        nullScript.commonPipelineEnvironment.configuration['steps']['whitesourceExecuteScan'] = nullScript.commonPipelineEnvironment.configuration['steps']['whitesourceExecuteScan'] ?: [:]
        nullScript.commonPipelineEnvironment.configuration['general'] = nullScript.commonPipelineEnvironment.configuration['general'] ?: [:]
        nullScript.commonPipelineEnvironment.configuration['general']['whitesource'] = nullScript.commonPipelineEnvironment.configuration['general']['whitesource'] ?: [:]
        nullScript.commonPipelineEnvironment.configuration['general']['whitesource']['serviceUrl'] = "http://some.host.whitesource.com/api/"
        nullScript.commonPipelineEnvironment.configuration['general']['whitesource']['userTokenCredentialsId'] = 'ID-123456789'
        nullScript.commonPipelineEnvironment.configuration['steps']['whitesourceExecuteScan']['userTokenCredentialsId'] = 'ID-123456789'
    }

    @Test
    void testMaven() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "1.0.0"
            ])
            return result
        })

        def publishHtmlMap = [:]
        helper.registerAllowedMethod("publishHTML", [Map.class], { m ->
            publishHtmlMap = m
            return null
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'maven',
            juStabUtils                          : utils,
            orgToken                             : 'testOrgToken',
            whitesourceProductName               : 'testProduct'
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'maven:3.5-jdk-8'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/java'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProduct\'')
        ))

        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProduct'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
    }


    @Test
    void testNpm() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "devDep": "false",
                "checkPolicies": "true",
                "projectName": "pipeline-test-node",
                "projectVersion": "1.0.0"
            ])
            return result
        })
        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'npm',
            juStabUtils                          : utils,
            orgToken                             : 'testOrgToken',
            productName                          : 'testProductName',
            productToken                         : 'testProductToken',
            reporting                            : false
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'node:8-stretch'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/node'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))
        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('curl --location --output jvm.tar.gz https://github.com/SAP/SapMachine/releases/download/sapmachine-11.0.2/sapmachine-jre-11.0.2_linux-x64_bin.tar.gz && tar --strip-components=1 -xzf jvm.tar.gz'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\'')
        ))

        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProductName'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productToken=testProductToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=com.sap.node.test-node'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
    }

    @Test
    void testNpmWithCustomConfigAndFilePath() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "devDep": "false",
                "checkPolicies": "true",
                "projectName": "pipeline-test-node",
                "projectVersion": "1.0.0"
            ])
            return result
        })
        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'npm',
            productName                          : 'SHC - Piper',
            configFilePath                       : './../../testConfigPath',
            file                                 : 'package.json',
            juStabUtils                          : utils,
            orgToken                             : 'b39d1328-52e2-42e3-98f0-932709daf3f0'
        ])

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('curl --location --output jvm.tar.gz https://github.com/SAP/SapMachine/releases/download/sapmachine-11.0.2/sapmachine-jre-11.0.2_linux-x64_bin.tar.gz && tar --strip-components=1 -xzf jvm.tar.gz'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./../../testConfigPath.2766cacc0cf1449dd4034385f4a9f0a6fdb755cf\' -apiKey \'b39d1328-52e2-42e3-98f0-932709daf3f0\' -userKey \'token-0815\' -product \'SHC - Piper\'')
        ))
    }

    @Test
    void testPip() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "1.0.0"
            ])
            return result
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'pip',
            juStabUtils                          : utils,
            orgToken                             : 'testOrgToken',
            productName                          : 'testProductName',
            reporting                            : false
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'python:3.7.2-stretch'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/python'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('curl --location --output jvm.tar.gz https://github.com/SAP/SapMachine/releases/download/sapmachine-11.0.2/sapmachine-jre-11.0.2_linux-x64_bin.tar.gz && tar --strip-components=1 -xzf jvm.tar.gz'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\'')
        ))

        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProductName'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=test-python'))
    }

    @Test
    void testWithOrgAdminCredentials() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "1.0.0"
            ])
            return result
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'pip',
            juStabUtils                          : utils,
            orgToken                             : 'testOrgToken',
            productName                          : 'testProductName',
            orgAdminUserTokenCredentialsId       : 'ID-9876543',
            reporting                            : false
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'python:3.7.2-stretch'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/python'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('curl --location --output jvm.tar.gz https://github.com/SAP/SapMachine/releases/download/sapmachine-11.0.2/sapmachine-jre-11.0.2_linux-x64_bin.tar.gz && tar --strip-components=1 -xzf jvm.tar.gz'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\'')
        ))

        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProductName'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=test-python'))
    }

    @Test
    void testNoProjectNoCreation() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "1.0.0"
            ])
            return result
        })

        def errorCaught = false
        try {
            stepRule.step.whitesourceExecuteScan([
                script                           : nullScript,
                whitesourceRepositoryStub        : whitesourceStub,
                whitesourceOrgAdminRepositoryStub: whitesourceOrgAdminRepositoryStub,
                descriptorUtilsStub              : descriptorUtilsStub,
                scanType                         : 'pip',
                juStabUtils                      : utils,
                orgToken                         : 'testOrgToken',
                productName                      : 'testProductName',
                createProductFromPipeline        : false,
                orgAdminUserTokenCredentialsId   : 'ID-9876543',
                reporting                        : false
            ])
        } catch (e) {
            errorCaught = true
            assertThat(e, isA(AbortException.class))
            assertThat(e.getMessage(), is("[WhiteSource] Could not fetch/find requested product 'testProductName' and automatic creation has been disabled"))
        }

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(errorCaught, is(true))
    }
    
    @Test
    void testSbt() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "1.0.0"
            ])
            return result
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'sbt',
            juStabUtils                          : utils,
            productName                          : 'testProductName',
            orgToken                             : 'testOrgToken',
            reporting                            : false
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'hseeberger/scala-sbt:8u181_2.12.8_1.2.8'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/scala'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\'')
        ))

        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProductName'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=com.sap.sbt.test-scala'))
    }

    @Test
    void testGo() {
        nullScript.commonPipelineEnvironment.gitHttpsUrl = 'https://github.wdf.sap.corp/test/golang'

        helper.registerAllowedMethod("readFile", [Map.class], {
            map ->
                def path = 'test/resources/DescriptorUtils/go/' + map.file.substring(map.file.lastIndexOf('/') + 1, map.file.length())
                def descriptorFile = new File(path)
                if(descriptorFile.exists())
                    return descriptorFile.text
                else
                    return null
        })

        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "2.0.0"
            ])
            return result
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'golang',
            juStabUtils                          : utils,
            productName                          : 'testProductName',
            orgToken                             : 'testOrgToken',
            reporting                            : false,
            buildDescriptorFile                  : './myProject/glide.yaml'
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'golang:1.12-stretch'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/dep'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./myProject/wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\'')
        ))

        assertThat(writeFileRule.files['./myProject/wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./myProject/wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProductName'))
        assertThat(writeFileRule.files['./myProject/wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
        assertThat(writeFileRule.files['./myProject/wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1'))
        assertThat(writeFileRule.files['./myProject/wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=github.wdf.sap.corp/test/golang.myProject'))
    }

    @Test
    void testGoDefaults() {
        nullScript.commonPipelineEnvironment.gitHttpsUrl = 'https://github.wdf.sap.corp/test/golang'

        helper.registerAllowedMethod("readFile", [Map.class], {
            map ->
                def path = 'test/resources/DescriptorUtils/go/' + map.file.substring(map.file.lastIndexOf('/') + 1, map.file.length())
                def descriptorFile = new File(path)
                if(descriptorFile.exists())
                    return descriptorFile.text
                else
                    return null
        })

        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "2.0.0"
            ])
            return result
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'golang',
            juStabUtils                          : utils,
            productName                          : 'testProductName',
            orgToken                             : 'testOrgToken',
            reporting                            : false
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'golang:1.12-stretch'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/dep'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('stashContent', ['buildDescriptor', 'opensourceConfiguration', 'modified whitesource config d3aa80454919391024374ba46b4df082d15ab9a3']))

        assertThat(shellRule.shell, Matchers.hasItems(
            is('curl --location --output wss-unified-agent.jar https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/wss-unified-agent.jar'),
            is('./bin/java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\'')
        ))

        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('apiKey=testOrgToken'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productName=testProductName'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('userKey=token-0815'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=github.wdf.sap.corp/test/golang'))
    }


    @Test
    void testAgentNoDownloads() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "2.0.0"
            ])
            return result
        })
        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'maven',
            agentDownloadUrl                     : '',
            jreDownloadUrl                       : '',
            agentParameters                      : 'testParams',
            juStabUtils                          : utils,
            orgToken                             : 'testOrgToken',
            productName                          : 'testProductName'
        ])

        assertThat(shellRule.shell[0], is('java -jar wss-unified-agent.jar -c \'./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3\' -apiKey \'testOrgToken\' -userKey \'token-0815\' -product \'testProductName\' testParams'))
    }

    @Test
    void testMta() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "2.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}pom.xml") {
                return [new File('maven1/pom.xml'), new File('maven2/pom.xml')].toArray()
            }
            if (map.glob == "**${File.separator}package.json") {
                return [new File('npm1/package.json'), new File('npm2/package.json')].toArray()
            }
            if (map.glob == "**${File.separator}setup.py") {
                return [new File('pip/setup.py')].toArray()
            }
            return [].toArray()
        })

        def whitesourceCalls = []
        helper.registerAllowedMethod("whitesourceExecuteScan", [Map.class], { map ->
            whitesourceCalls.add(map)
            stepRule.step.call(map)
        })

        def parallelMap = [:]
        helper.registerAllowedMethod("parallel", [Map.class], { map ->
            parallelMap = map
            parallelMap['Whitesource - maven1']()
            parallelMap['Whitesource - npm1']()
            parallelMap['Whitesource - pip']()
        })

        //need to use call due to mock above
        stepRule.step.call([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'mta',
            whitesource: [
                productName                          : 'SHC - Piper',
                orgToken                             : 'b39d1328-52e2-42e3-98f0-932709daf3f0'
            ],
            buildDescriptorExcludeList           : ["maven2${File.separator}pom.xml".toString(), "npm2${File.separator}package.json".toString()],
            reporting                            : true,
            juStabUtils                          : utils
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(parallelMap, hasKey('Whitesource - maven1'))
        assertThat(parallelMap, hasKey('Whitesource - npm1'))
        assertThat(parallelMap, hasKey('Whitesource - pip'))
        assertThat(parallelMap.keySet(), hasSize(4))

        assertThat(whitesourceCalls,
            contains(
                allOf(
                    hasEntry('scanType', 'maven'),
                    hasEntry('buildDescriptorFile', "maven1${File.separator}pom.xml".toString())
                ),
                allOf(
                    hasEntry('scanType', 'npm'),
                    hasEntry('buildDescriptorFile', "npm1${File.separator}package.json".toString())
                ),
                allOf(
                    hasEntry('scanType', 'pip'),
                    hasEntry('buildDescriptorFile', "pip${File.separator}setup.py".toString())
                )
            )
        )

        assertThat(whitesourceCalls[0]['whitesource']['projectNames'], contains("com.sap.maven.test-java - 1", "com.sap.node.test-node - 1", "test-python - 1"))
        assertThat(whitesourceCalls[1]['whitesource']['projectNames'], contains("com.sap.maven.test-java - 1", "com.sap.node.test-node - 1", "test-python - 1"))
        assertThat(whitesourceCalls[2]['whitesource']['projectNames'], contains("com.sap.maven.test-java - 1", "com.sap.node.test-node - 1", "test-python - 1"))
    }

    @Test
    void testMtaBlocks() {
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}pom.xml") {
                return [new File('maven1/pom.xml'), new File('maven2/pom.xml')].toArray()
            }
            if (map.glob == "**${File.separator}package.json") {
                return [new File('npm1/package.json'), new File('npm2/package.json'), new File('npm3/package.json'), new File('npm4/package.json')].toArray()
            }
            return [].toArray()
        })

        def whitesourceCalls = []
        helper.registerAllowedMethod("whitesourceExecuteScan", [Map.class], { map ->
            whitesourceCalls.add(map)
        })

        def invocation = 0
        def parallelMap = [:]
        helper.registerAllowedMethod("parallel", [Map.class], { map ->
            parallelMap[invocation] = map
            for(i = 0; i < parallelMap[invocation].keySet().size(); i++) {
                def key = parallelMap[invocation].keySet()[i]
                if(key != "failFast")
                    parallelMap[invocation][key]()
            }
            invocation++
        })

        //need to use call due to mock above
        stepRule.step.call([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            scanType                             : 'mta',
            productName                          : 'SHC - Piper',
            buildDescriptorExcludeList           : ["maven2${File.separator}pom.xml".toString()],
            juStabUtils                          : utils,
            parallelLimit                        : 3,
            orgToken                             : 'b39d1328-52e2-42e3-98f0-932709daf3f0'
        ])

        assertThat(loggingRule.log, containsString('Unstash content: buildDescriptor'))
        assertThat(loggingRule.log, containsString('Unstash content: opensourceConfiguration'))

        assertThat(invocation, is(2))
        assertThat(parallelMap[0], hasKey('Whitesource - maven1'))
        assertThat(parallelMap[0], hasKey('Whitesource - npm1'))
        assertThat(parallelMap[0], hasKey('Whitesource - npm2'))
        assertThat(parallelMap[0].keySet(), hasSize(4))
        assertThat(parallelMap[1], hasKey('Whitesource - npm3'))
        assertThat(parallelMap[1], hasKey('Whitesource - npm4'))
        assertThat(parallelMap[1].keySet(), hasSize(3))

        assertThat(whitesourceCalls, hasItem(allOf(
            hasEntry('scanType', 'maven'),
            hasEntry('buildDescriptorFile', "maven1${File.separator}pom.xml".toString())
        )))
        assertThat(whitesourceCalls, hasItem(allOf(
            hasEntry('scanType', 'npm'),
            hasEntry('buildDescriptorFile', "npm1${File.separator}package.json".toString())
        )))
        assertThat(whitesourceCalls, hasItem(allOf(
            hasEntry('scanType', 'npm'),
            hasEntry('buildDescriptorFile', "npm2${File.separator}package.json".toString())
        )))
        assertThat(whitesourceCalls, hasItem(allOf(
            hasEntry('scanType', 'npm'),
            hasEntry('buildDescriptorFile', "npm3${File.separator}package.json".toString())
        )))
        assertThat(whitesourceCalls, hasItem(allOf(
            hasEntry('scanType', 'npm'),
            hasEntry('buildDescriptorFile', "npm4${File.separator}package.json".toString())
        )))
    }

    @Test
    void testMtaSingleExclude() {

        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}pom.xml") {
                return [new File('maven1/pom.xml'), new File('maven2/pom.xml')].toArray()
            }
            if (map.glob == "**${File.separator}package.json") {
                return [new File('npm1/package.json'), new File('npm2/package.json')].toArray()
            }
            return [].toArray()
        })

        def parallelMap = [:]
        helper.registerAllowedMethod("parallel", [Map.class], { map ->
            parallelMap = map
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            scanType                             : 'mta',
            productName                          : 'SHC - Piper',
            buildDescriptorExcludeList           : "maven2${File.separator}pom.xml",
            juStabUtils                          : utils,
            orgToken                             : 'b39d1328-52e2-42e3-98f0-932709daf3f0'
        ])

        assertThat(parallelMap.keySet(), hasSize(4))

    }

    @Test
    void testNPMStatusCheckScanException() {
        thrown.expect(AbortException.class)
        stepRule.step.checkStatus(-1 & 0xFF, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testNPMStatusCheckPolicyViolation() {
        thrown.expect(AbortException.class)
        stepRule.step.checkStatus(-2 & 0xFF, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testNPMStatusCheckNoPolicyViolation() {
        stepRule.step.checkStatus(-2 & 0xFF, [whitesource:[licensingVulnerabilities: false]])
    }

    @Test
    void testNPMStatusCheckClientException() {
        thrown.expect(AbortException.class)
        stepRule.step.checkStatus(-3 & 0xFF, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testNPMStatusCheckConnectionException() {
        thrown.expect(AbortException.class)
        stepRule.step.checkStatus(-4 & 0xFF, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testNPMStatusCheckServerException() {
        thrown.expect(AbortException.class)
        stepRule.step.checkStatus(-3 & 0xFF, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testViolationPolicy() {
        stepRule.step.checkViolationStatus(0)
    }

    @Test
    void testViolationPolicyException() {
        thrown.expect(AbortException.class)
        thrown.expectMessage("1")
        stepRule.step.checkViolationStatus(1)
    }

    @Test
    void testFetchViolationCountProject() {

        def config = [projectNames: ["piper-java-cc - 0.0.1", "pipeline-test - 0.0.1"], productToken: "product-token"]

        def projectTokens =  [[token:"abc-project-token"],[token:"def-project-token"]]
        helper.registerAllowedMethod('fetchProjectsMetaInfo', [], { return projectTokens })
        helper.registerAllowedMethod('fetchProjectLicenseAlerts', [String], { projectToken ->
            if(projectToken == projectTokens[0].token)
                return [alerts: []]
            if(projectToken == projectTokens[1].token)
                return [alerts: []]
        })

        def violationCount = stepRule.step.fetchViolationCount(config, whitesourceStub)

        Assert.assertTrue(violationCount == 0)
    }


    @Test
    void  testFetchViolationCountProjectNotZero() {

        def config = [whitesource: [projectNames: ["piper-java-cc - 0.0.1", "pipeline-test - 0.0.1"]]]
        def projectTokens = [[token:"abc-project-token"],[token:"def-project-token"]]
        helper.registerAllowedMethod('fetchProjectsMetaInfo', [], { return projectTokens })
        helper.registerAllowedMethod('fetchProjectLicenseAlerts', [String], { projectToken ->
            if(projectToken == projectTokens[0].token)
                return [alerts: [{}, {}]]
            if(projectToken == projectTokens[1].token)
                return [alerts: [{}, {}, {}]]
        })

        def violationCount = stepRule.step.fetchViolationCount(config, whitesourceStub)

        Assert.assertTrue(violationCount == 5)
    }

    @Test
    void testFetchViolationCountProduct() {

        def config = [:]

        helper.registerAllowedMethod('fetchProductLicenseAlerts', [], { return [alerts: []] })

        def violationCount = stepRule.step.fetchViolationCount(config, whitesourceStub)

        Assert.assertTrue(violationCount == 0)
    }

    @Test
    void testFetchViolationCountProductNotZero() {

        def config = [:]

        helper.registerAllowedMethod('fetchProductLicenseAlerts', [], { return [alerts: [{}, {}]]})

        def violationCount = stepRule.step.fetchViolationCount(config, whitesourceStub)

        Assert.assertTrue(violationCount == 2)
    }

    @Test
    void testCheckException() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "python-test",
                "projectVersion": "2.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("""{"alerts":[{"vulnerability":{"name":"CVE-2017-15095","type":"CVE","severity":"high","score":7.5,"cvss3_severity":"high","cvss3_score":9.8,"scoreMetadataVector":"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H","publishDate":"2018-02-06","url":"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-15095",
                                                                    "description":"A deserialization flaw was discovered in the jackson-databind in versions before 2.8.10 and 2.9.1, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper. This issue extends the previous flaw CVE-2017-7525 by blacklisting more classes that could be used maliciously.",
                                                                    "topFix":{"vulnerability":"CVE-2017-15095","type":"CHANGE_FILES","origin":"GITHUB_COMMIT","url":"https://github.com/FasterXML/jackson-databind/commit/60d459ce","fixResolution":"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java","date":"2017-04-13",
                                                                    "message":"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8","extraData":"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4"},"allFixes":[{"vulnerability":"CVE-2017-15095","type":"CHANGE_FILES","origin":"GITHUB_COMMIT",
                                                                    "url":"https://github.com/FasterXML/jackson-databind/commit/60d459ce","fixResolution":"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java","date":"2017-04-13","message":"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8",
                                                                    "extraData":"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4"},{"vulnerability":"CVE-2017-15095","type":"CHANGE_FILES","origin":"GITHUB_COMMIT","url":"https://github.com/FasterXML/jackson-databind/commit/e865a7a4464da63ded9f4b1a2328ad85c9ded78b#diff-98084d808198119d550a9211e128a16f",
                                                                    "fixResolution":"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java","date":"2017-12-12","message":"Fix #1737 (#1857)","extraData":"key=e865a7a&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4"},
                                                                    {"vulnerability":"CVE-2017-15095","type":"CHANGE_FILES","origin":"GITHUB_COMMIT","url":"https://github.com/FasterXML/jackson-databind/commit/e8f043d1","fixResolution":"release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java","date":"2017-06-30","message":"Fix #1680",
                                                                    "extraData":"key=e8f043d&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4"}],"fixResolutionText":"Replace or update the following files: IllegalTypesCheckTest.java, VERSION, BeanDeserializerFactory.java","references":[]},
                                                                    "type":"SECURITY_VULNERABILITY","level":"MAJOR","library":{"keyUuid":"13f7802e-8aa1-4303-a5db-1d0c85e871a9","keyId":23410061,"filename":"jackson-databind-2.8.8.jar","name":"jackson-databind","groupId":"com.fasterxml.jackson.core","artifactId":"jackson-databind","version":"2.8.8","sha1":"bf88c7b27e95cbadce4e7c316a56c3efffda8026",
                                                                    "type":"Java","references":{"url":"http://github.com/FasterXML/jackson","issueUrl":"https://github.com/FasterXML/jackson-databind/issues","pomUrl":"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom","scmUrl":"http://github.com/FasterXML/jackson-databind"},
                                                                    "licenses":[{"name":"Apache 2.0","url":"http://apache.org/licenses/LICENSE-2.0","profileInfo":{"copyrightRiskScore":"THREE","patentRiskScore":"ONE","copyleft":"NO","linking":"DYNAMIC","royaltyFree":"CONDITIONAL"}}]},"project":"pipeline-test - 0.0.1","projectId":302194,"projectToken":"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7","directDependency":false,"description":"High:5,","date":"2017-11-15"}]}""").alerts
        })

        thrown.expect(AbortException.class)

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            scanType                             : 'npm',
            juStabUtils                          : utils,
            securityVulnerabilities              : true,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            cvssSeverityLimit                    : 7,
            orgToken                             : 'testOrgToken',
            productName                          : 'SHC - Piper',
            projectNames                         : [ 'piper-demo - 0.0.1' ]
        ])
    }

    @Test
    void testCheckFindingBelowThreshold() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "pipeline-test-node",
                "projectVersion": "2.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("""{ \"alerts\": [ { \"vulnerability\": { \"name\": \"CVE-2017-15095\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 2.1, \"cvss3_severity\": \"high\", \"cvss3_score\": 5.3, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-06\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-15095\", \"description\": \"A deserialization flaw was discovered in the jackson-databind in versions before 2.8.10 and 2.9.1, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper. This issue extends the previous flaw CVE-2017-7525 by blacklisting more classes that could be used maliciously.\", \"topFix\": { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\",
                \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\", \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\",
                \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e865a7a4464da63ded9f4b1a2328ad85c9ded78b#diff-98084d808198119d550a9211e128a16f\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-12-12\", \"message\": \"Fix #1737 (#1857)\", \"extraData\": \"key=e865a7a&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\",
                \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e8f043d1\", \"fixResolution\": \"release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-06-30\", \"message\": \"Fix #1680\", \"extraData\": \"key=e8f043d&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: IllegalTypesCheckTest.java, VERSION, BeanDeserializerFactory.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\",
                \"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" } ] }""").alerts
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'npm',
            juStabUtils                          : utils,
            securityVulnerabilities              : true,
            orgToken                             : 'testOrgToken',
            productName                          : 'SHC - Piper',
            projectNames                         : [ 'piper-demo - 0.0.1' ],
            cvssSeverityLimit                    : 7
        ])

        assertThat(loggingRule.log, containsString('WARNING: 1 Open Source Software Security vulnerabilities with CVSS score below 7 detected.'))
        assertThat(writeFileRule.files['piper_whitesource_vulnerability_report.json'], not(isEmptyOrNullString()))
    }

    @Test
    void testCheckFindingAbove() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "pipeline-test-node",
                "projectVersion": "2.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("""{ \"alerts\": [ { \"vulnerability\": { \"name\": \"CVE-2017-15095\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 2.1, \"cvss3_severity\": \"high\", \"cvss3_score\": 5.3, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-06\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-15095\", \"description\": \"A deserialization flaw was discovered in the jackson-databind in versions before 2.8.10 and 2.9.1, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper. This issue extends the previous flaw CVE-2017-7525 by blacklisting more classes that could be used maliciously.\", \"topFix\": { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\",
                \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\", \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\",
                \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e865a7a4464da63ded9f4b1a2328ad85c9ded78b#diff-98084d808198119d550a9211e128a16f\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-12-12\", \"message\": \"Fix #1737 (#1857)\", \"extraData\": \"key=e865a7a&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\",
                \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e8f043d1\", \"fixResolution\": \"release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-06-30\", \"message\": \"Fix #1680\", \"extraData\": \"key=e8f043d&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: IllegalTypesCheckTest.java, VERSION, BeanDeserializerFactory.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\",
                \"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" } ] }""").alerts
        })

        thrown.expect(AbortException)
        thrown.expectMessage('[whitesourceExecuteScan] 1 Open Source Software Security vulnerabilities with CVSS score greater or equal 0 detected. - ')

        stepRule.step.whitesourceExecuteScan([
            script                           : nullScript,
            whitesourceRepositoryStub        : whitesourceStub,
            whitesourceOrgAdminRepositoryStub: whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub              : descriptorUtilsStub,
            scanType                         : 'npm',
            juStabUtils                      : utils,
            securityVulnerabilities          : true,
            orgToken                         : 'testOrgToken',
            productName                      : 'SHC - Piper',
            projectNames                     : ['piper-demo - 0.0.1'],
            cvssSeverityLimit                : 0
        ])

        assertThat(writeFileRule.files['piper_whitesource_vulnerability_report.json'], not(isEmptyOrNullString()))
    }

    @Test
    void testCheckNoFindings() {
        helper.registerAllowedMethod("readProperties", [Map], {
            def result = new Properties()
            result.putAll([
                "apiKey": "b39d1328-52e2-42e3-98f0-932709daf3f0",
                "productName": "SHC - Piper",
                "checkPolicies": "true",
                "projectName": "pipeline-test-node",
                "projectVersion": "2.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [Object.class], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [] }").alerts
        })

        stepRule.step.whitesourceExecuteScan([
            script                               : nullScript,
            whitesourceRepositoryStub            : whitesourceStub,
            whitesourceOrgAdminRepositoryStub    : whitesourceOrgAdminRepositoryStub,
            descriptorUtilsStub                  : descriptorUtilsStub,
            scanType                             : 'npm',
            juStabUtils                          : utils,
            securityVulnerabilities              : true,
            orgToken                             : 'testOrgToken',
            productName                          : 'SHC - Piper',
            projectNames                         : [ 'piper-demo - 0.0.1' ]
        ])

        assertThat(loggingRule.log, containsString('No Open Source Software Security vulnerabilities detected.'))
        assertThat(writeFileRule.files['piper_whitesource_vulnerability_report.json'], not(isEmptyOrNullString()))
    }

    @Test
    void testCheckStatus_0() {
        stepRule.step.checkStatus(0, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_255() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] The scan resulted in an error")
        stepRule.step.checkStatus(255, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_254() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] Whitesource found one or multiple policy violations")
        stepRule.step.checkStatus(254, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_253() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] The local scan client failed to execute the scan")
        stepRule.step.checkStatus(253, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_252() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] There was a failure in the connection to the WhiteSource servers")
        stepRule.step.checkStatus(252, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_251() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] The server failed to analyze the scan")
        stepRule.step.checkStatus(251, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_250() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] Pre-step failure")
        stepRule.step.checkStatus(250, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_127() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] Whitesource scan failed with unknown error code '127'")
        stepRule.step.checkStatus(127, [whitesource:[licensingVulnerabilities: true]])
    }

    @Test
    void testCheckStatus_vulnerability() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] 5 Open Source Software Security vulnerabilities with CVSS score greater or equal 7 detected. - ")
        stepRule.step.checkStatus(0, [whitesource:[licensingVulnerabilities: false, securityVulnerabilities: true, severeVulnerabilities: 5, cvssSeverityLimit: 7]])
    }

    @Test
    void testCheckViolationStatus_0() {
        stepRule.step.checkViolationStatus(0)
        assertThat(loggingRule.log, containsString ("[whitesourceExecuteScan] No policy violations found"))
    }

    @Test
    void testCheckViolationStatus_5() {
        thrown.expect(AbortException)
        thrown.expectMessage("[whitesourceExecuteScan] Whitesource found 5 policy violations for your product")
        stepRule.step.checkViolationStatus(5)
    }
}
