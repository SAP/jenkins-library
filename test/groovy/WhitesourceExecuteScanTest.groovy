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
    private JenkinsErrorRule errorRule = new JenkinsErrorRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(environmentRule)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(loggingRule)
        .around(writeFileRule)
        .around(stepRule)
        .around(errorRule)

    def whitesourceOrgAdminRepositoryStub
    def whitesourceStub

    @Autowired
    DescriptorUtils descriptorUtilsStub

    @Before
    void init() {
        def credentialsStore = ['ID-123456789': 'token-0815', 'ID-9876543': 'token-0816', 'ID-abcdefg': ['testUser', 'testPassword']]
        def withCredentialsBindings
        helper.registerAllowedMethod('string', [Map], {
            m ->
                withCredentialsBindings = ["${m.credentialsId}": "${m.variable}"]
                return m
        })
        helper.registerAllowedMethod('usernamePassword', [Map], {
            m ->
                withCredentialsBindings = ["${m.credentialsId}": ["${m.usernameVariable}", "${m.passwordVariable}"]]
                return m
        })
        helper.registerAllowedMethod('withCredentials', [List.class, Closure.class], {
            l, body ->
                def index = 0
                withCredentialsBindings.each {
                    entry ->
                        if(entry.value instanceof List) {
                            entry.value.each {
                                subEntry ->
                                    def value = credentialsStore[entry.key]
                                    getBinding().setProperty(subEntry, value[index])
                                    index++

                            }
                        } else {
                            getBinding().setProperty(entry.value, credentialsStore[entry.key])
                        }
                }
                try {
                    body()
                } finally {
                    withCredentialsBindings.each {
                        entry ->
                            if(entry.value instanceof List) {
                                entry.value.each {
                                    subEntry ->
                                        getBinding().setProperty(subEntry, null)

                                }
                            } else {
                                getBinding().setProperty(entry.value, null)
                            }
                    }
                }
        })
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
        nullScript.commonPipelineEnvironment.setConfigProperty('userTokenCredentialsId', 'ID-123456789')
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
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1.2.3'))
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
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1.2.3'))
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
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1.2.3'))
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
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('productVersion=1.2.3'))
        assertThat(writeFileRule.files['./wss-unified-agent.config.d3aa80454919391024374ba46b4df082d15ab9a3'], containsString('projectName=com.sap.sbt.test-scala'))
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
                "projectVersion": "1.0.0"
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
                "projectVersion": "1.0.0"
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

        assertThat(whitesourceCalls[0]['whitesource']['projectNames'], contains("com.sap.maven.test-java - 1.2.3", "com.sap.node.test-node - 1.2.3", "test-python - 1.2.3"))
        assertThat(whitesourceCalls[1]['whitesource']['projectNames'], contains("com.sap.maven.test-java - 1.2.3", "com.sap.node.test-node - 1.2.3", "test-python - 1.2.3"))
        assertThat(whitesourceCalls[2]['whitesource']['projectNames'], contains("com.sap.maven.test-java - 1.2.3", "com.sap.node.test-node - 1.2.3", "test-python - 1.2.3"))
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
                "projectVersion": "1.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [ { \"vulnerability\": { \"name\": \"CVE-2017-15095\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-06\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-15095\", \"description\": \"A deserialization flaw was discovered in the jackson-databind in versions before 2.8.10 and 2.9.1, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper. This issue extends the previous flaw CVE-2017-7525 by blacklisting more classes that could be used maliciously.\", \"topFix\": { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\"," +
                "\"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\", \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\"," +
                "\"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e865a7a4464da63ded9f4b1a2328ad85c9ded78b#diff-98084d808198119d550a9211e128a16f\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-12-12\", \"message\": \"Fix #1737 (#1857)\", \"extraData\": \"key=e865a7a&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e8f043d1\"," +
                "\"fixResolution\": \"release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-06-30\", \"message\": \"Fix #1680\", \"extraData\": \"key=e8f043d&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: IllegalTypesCheckTest.java, VERSION, BeanDeserializerFactory.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\", \"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\"," +
                "\"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" }, { \"vulnerability\": { \"name\": \"CVE-2017-17485\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-01-10\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-17485\"," +
                "\"description\": \"FasterXML jackson-databind through 2.8.10 and 2.9.x through 2.9.3 allows unauthenticated remote code execution because of an incomplete fix for the CVE-2017-7525 deserialization flaw. This is exploitable by sending maliciously crafted JSON input to the readValue method of the ObjectMapper, bypassing a blacklist that is ineffective if the Spring libraries are available in the classpath.\", \"topFix\": { \"vulnerability\": \"CVE-2017-17485\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/bb45fb16709018842f858f1a6e1118676aaa34bd#diff-727a6e8db3603b95f185697108af6c48\", \"fixResolution\": \"src/test/java/org/springframework/jacksontest/AbstractApplicationContext.java,src/test/java/org/springframework/jacksontest/AbstractPointcutAdvisor.java,src/test/java/org/springframework/jacksontest/BogusApplicationContext.java,src/main/java/com/fasterxml/jackson/databind/jsontype/impl/SubTypeValidator.java,src/test/java/org/springframework/jacksontest/BogusPointcutAdvisor.java,src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java\"," +
                "\"date\": \"2017-12-19\", \"message\": \"Fix issues with earlier fix for #1855\", \"extraData\": \"key=bb45fb1&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-17485\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/bb45fb16709018842f858f1a6e1118676aaa34bd#diff-727a6e8db3603b95f185697108af6c48\", \"fixResolution\": \"src/test/java/org/springframework/jacksontest/AbstractApplicationContext.java,src/test/java/org/springframework/jacksontest/AbstractPointcutAdvisor.java,src/test/java/org/springframework/jacksontest/BogusApplicationContext.java,src/main/java/com/fasterxml/jackson/databind/jsontype/impl/SubTypeValidator.java,src/test/java/org/springframework/jacksontest/BogusPointcutAdvisor.java,src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java\", \"date\": \"2017-12-19\", \"message\": \"Fix issues with earlier fix for #1855\"," +
                "\"extraData\": \"key=bb45fb1&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-17485\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/bb45fb16709018842f858f1a6e1118676aaa34bd\", \"fixResolution\": \"src/test/java/org/springframework/jacksontest/AbstractApplicationContext.java,src/test/java/org/springframework/jacksontest/AbstractPointcutAdvisor.java,src/test/java/org/springframework/jacksontest/BogusApplicationContext.java,src/main/java/com/fasterxml/jackson/databind/jsontype/impl/SubTypeValidator.java,src/test/java/org/springframework/jacksontest/BogusPointcutAdvisor.java,src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java\", \"date\": \"2017-12-19\", \"message\": \"Fix issues with earlier fix for #1855\", \"extraData\": \"key=bb45fb1&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ]," +
                "\"fixResolutionText\": \"Replace or update the following files: AbstractApplicationContext.java, AbstractPointcutAdvisor.java, BogusApplicationContext.java, SubTypeValidator.java, BogusPointcutAdvisor.java, IllegalTypesCheckTest.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\", \"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\"," +
                "\"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" }, { \"vulnerability\": { \"name\": \"CVE-2017-7525\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-06\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-7525\", \"description\": \"A deserialization flaw was discovered in the jackson-databind, versions before 2.6.7.1, 2.7.9.1 and 2.8.9, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper.\", \"topFix\": { \"vulnerability\": \"CVE-2017-7525\"," +
                "\"type\": \"UPGRADE_VERSION\", \"origin\": \"BUGZILLA\", \"url\": \"https://bugzilla.redhat.com/show_bug.cgi?id=CVE-2017-7525\", \"fixResolution\": \"jackson-databind 2.8.9,jackson-databind 2.9.0\", \"message\": \"CVE-2017-7525 jackson-databind: Deserialization vulnerability via readValue method of ObjectMapper\", \"extraData\": \"key=1462702&assignee=Red Hat Product Security\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-7525\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"BUGZILLA\", \"url\": \"https://bugzilla.redhat.com/show_bug.cgi?id=CVE-2017-7525\", \"fixResolution\": \"jackson-databind 2.8.9,jackson-databind 2.9.0\", \"message\": \"CVE-2017-7525 jackson-databind: Deserialization vulnerability via readValue method of ObjectMapper\", \"extraData\": \"key=1462702&assignee=Red Hat Product Security\" }, { \"vulnerability\": \"CVE-2017-7525\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459cedcf079c6106ae7da2ac562bc32dcabe1#diff-98084d808198119d550a9211e128a16f\"," +
                "\"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\", \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-7525\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459cedcf079c6106ae7da2ac562bc32dcabe1\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\"," +
                "\"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-7525\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/6ce32ffd18facac6abdbbf559c817b47fcb622c1\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.7(.10)\", \"extraData\": \"key=6ce32ff&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ], \"fixResolutionText\": \"Upgrade to version jackson-databind 2.8.9, jackson-databind 2.9.0 or greater\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\"," +
                "\"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\", \"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }," +
                "\"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" }, { \"vulnerability\": { \"name\": \"CVE-2018-5968\", \"type\": \"CVE\", \"severity\": \"medium\", \"score\": 5.1, \"cvss3_severity\": \"high\", \"cvss3_score\": 8.1, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-01-22\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-5968\", \"description\": \"FasterXML jackson-databind through 2.8.11 and 2.9.x through 2.9.3 allows unauthenticated remote code execution because of an incomplete fix for the CVE-2017-7525 and CVE-2017-17485 deserialization flaws. This is exploitable via two different gadgets that bypass a blacklist.\", \"topFix\": { \"vulnerability\": \"CVE-2018-5968\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\"," +
                "\"url\": \"https://github.com/FasterXML/jackson-databind/commit/038b471e2efde2e8f96b4e0be958d3e5a1ff1d05\", \"fixResolution\": \"src/main/java/com/fasterxml/jackson/databind/jsontype/impl/SubTypeValidator.java,release-notes/VERSION\", \"date\": \"2018-01-22\", \"message\": \"Fix #1899\", \"extraData\": \"key=038b471&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [], \"fixResolutionText\": \"Replace or update the following files: SubTypeValidator.java, VERSION\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\"," +
                "\"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" }, { \"vulnerability\": { \"name\": \"CVE-2018-7489\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5," +
                "\"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-26\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-7489\", \"description\": \"FasterXML jackson-databind before 2.8.11.1 and 2.9.x before 2.9.5 allows unauthenticated remote code execution because of an incomplete fix for the CVE-2017-7525 deserialization flaw. This is exploitable by sending maliciously crafted JSON input to the readValue method of the ObjectMapper, bypassing a blacklist that is ineffective if the c3p0 libraries are available in the classpath.\", \"topFix\": { \"vulnerability\": \"CVE-2018-7489\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/6799f8f10cc78e9af6d443ed6982d00a13f2e7d2\", \"fixResolution\": \"src/main/java/com/fasterxml/jackson/databind/jsontype/impl/SubTypeValidator.java,src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,src/test/java/com/mchange/v2/c3p0/jacksontest/ComboPooledDataSource.java,release-notes/VERSION\"," +
                "\"date\": \"2018-02-11\", \"message\": \"Fix #1931\", \"extraData\": \"key=6799f8f&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [], \"fixResolutionText\": \"Replace or update the following files: SubTypeValidator.java, IllegalTypesCheckTest.java, ComboPooledDataSource.java, VERSION\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\", \"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\"," +
                "\"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" }, { \"vulnerability\": { \"name\": \"CVE-2016-3720\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2016-06-10\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2016-3720\"," +
                "\"description\": \"XML external entity (XXE) vulnerability in XmlMapper in the Data format extension for Jackson (aka jackson-dataformat-xml) allows attackers to have unspecified impact via unknown vectors.\", \"allFixes\": [], \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"9c7e7805-0ba0-480a-8647-ed0a457b4a88\", \"keyId\": 441802, \"filename\": \"jackson-dataformat-xml-2.4.2.jar\", \"name\": \"Jackson-dataformat-XML\", \"groupId\": \"com.fasterxml.jackson.dataformat\", \"artifactId\": \"jackson-dataformat-xml\", \"version\": \"2.4.2\", \"sha1\": \"02f2d96f68b2d3475452d95dde7a3fbee225f6ae\", \"type\": \"Java\", \"references\": { \"url\": \"http://wiki.fasterxml.com/JacksonExtensionXmlDataBinding\", \"issueUrl\": \"https://github.com/FasterXML/jackson-dataformat-xml/issues\", \"pomUrl\": \"http://maven.ibiblio.org/maven2/com/fasterxml/jackson/dataformat/jackson-dataformat-xml/2.4.2/jackson-dataformat-xml-2.4.2.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-dataformat-xml\" }," +
                "\"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:1,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2017-5929\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2017-03-13\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-5929\", \"description\": \"QOS.ch Logback before 1.2.0 has a serialization vulnerability affecting the SocketServer and ServerSocketReceiver components.\", \"topFix\": { \"vulnerability\": \"CVE-2017-5929\"," +
                "\"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/qos-ch/logback/commit/f46044b805bca91efe5fd6afe52257cd02f775f8\", \"fixResolution\": \"logback-classic/src/main/java/ch/qos/logback/classic/net/server/LogbackClassicSerializationHelper.java,logback-core/src/test/java/ch/qos/logback/core/net/Innocent.java,logback-classic/src/main/java/ch/qos/logback/classic/net/SimpleSocketServer.java,logback-core/src/main/java/ch/qos/logback/core/net/HardenedObjectInputStream.java,logback-classic/src/test/java/ch/qos/logback/classic/LoggerSerializationTest.java,logback-core/src/test/java/ch/qos/logback/core/net/HardenedObjectInputStreamTest.java\", \"date\": \"2017-02-07\", \"message\": \"harden serialization\", \"extraData\": \"key=f46044b&committerName=ceki&committerUrl=https://github.com/ceki&committerAvatar=https://avatars1.githubusercontent.com/u/115476?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-5929\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\"," +
                "\"url\": \"https://github.com/qos-ch/logback/commit/f46044b805bca91efe5fd6afe52257cd02f775f8\", \"fixResolution\": \"logback-classic/src/main/java/ch/qos/logback/classic/net/server/LogbackClassicSerializationHelper.java,logback-core/src/test/java/ch/qos/logback/core/net/Innocent.java,logback-classic/src/main/java/ch/qos/logback/classic/net/SimpleSocketServer.java,logback-core/src/main/java/ch/qos/logback/core/net/HardenedObjectInputStream.java,logback-classic/src/test/java/ch/qos/logback/classic/LoggerSerializationTest.java,logback-core/src/test/java/ch/qos/logback/core/net/HardenedObjectInputStreamTest.java\", \"date\": \"2017-02-07\", \"message\": \"harden serialization\", \"extraData\": \"key=f46044b&committerName=ceki&committerUrl=https://github.com/ceki&committerAvatar=https://avatars1.githubusercontent.com/u/115476?v=4\" }, { \"vulnerability\": \"CVE-2017-5929\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/victims/victims-cve-db/commit/94745e07d4b10a58d98ffc3916d9711fa407b018\"," +
                "\"fixResolution\": \"database/java/2017/5929.yaml\", \"date\": \"2017-03-15\", \"message\": \"Added CVE-2017-5929 per issue #76\", \"extraData\": \"key=94745e0&committerName=cplvic&committerUrl=https://github.com/cplvic&committerAvatar=https://avatars0.githubusercontent.com/u/11528385?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: LogbackClassicSerializationHelper.java, Innocent.java, SimpleSocketServer.java, HardenedObjectInputStream.java, LoggerSerializationTest.java, HardenedObjectInputStreamTest.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"1259a151-8d7c-4921-b293-3b4b8708ac5e\", \"keyId\": 754707, \"filename\": \"logback-classic-1.1.3.jar\", \"name\": \"Logback Classic Module\", \"groupId\": \"ch.qos.logback\", \"artifactId\": \"logback-classic\", \"version\": \"1.1.3\", \"sha1\": \"d90276fff414f06cb375f2057f6778cd63c6082f\", \"type\": \"Java\", \"references\": { \"url\": \"http://logback.qos.ch/logback-classic\"," +
                "\"pomUrl\": \"http://maven.ibiblio.org/maven2/ch/qos/logback/logback-classic/1.1.3/logback-classic-1.1.3.pom\", \"scmUrl\": \"https://github.com/ceki/logback/logback-classic\" }, \"licenses\": [ { \"name\": \"LGPL 2.1\", \"url\": \"http://opensource.org/licenses/lgpl-2.1\", \"profileInfo\": { \"copyrightRiskScore\": \"FIVE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"PARTIAL\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } }, { \"name\": \"Eclipse 1.0\", \"url\": \"http://opensource.org/licenses/eclipse-1.0.php\", \"profileInfo\": { \"copyrightRiskScore\": \"SIX\", \"patentRiskScore\": \"FOUR\", \"copyleft\": \"PARTIAL\", \"linking\": \"VIRAL\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": true, \"description\": \"High:1,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2017-5929\", \"type\": \"CVE\", \"severity\": \"high\"," +
                "\"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2017-03-13\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-5929\", \"description\": \"QOS.ch Logback before 1.2.0 has a serialization vulnerability affecting the SocketServer and ServerSocketReceiver components.\", \"topFix\": { \"vulnerability\": \"CVE-2017-5929\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/qos-ch/logback/commit/f46044b805bca91efe5fd6afe52257cd02f775f8\", \"fixResolution\": \"logback-classic/src/main/java/ch/qos/logback/classic/net/server/LogbackClassicSerializationHelper.java,logback-core/src/test/java/ch/qos/logback/core/net/Innocent.java,logback-classic/src/main/java/ch/qos/logback/classic/net/SimpleSocketServer.java,logback-core/src/main/java/ch/qos/logback/core/net/HardenedObjectInputStream.java,logback-classic/src/test/java/ch/qos/logback/classic/LoggerSerializationTest.java,logback-core/src/test/java/ch/qos/logback/core/net/HardenedObjectInputStreamTest.java\", \"date\": \"2017-02-07\"," +
                "\"message\": \"harden serialization\", \"extraData\": \"key=f46044b&committerName=ceki&committerUrl=https://github.com/ceki&committerAvatar=https://avatars1.githubusercontent.com/u/115476?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-5929\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/qos-ch/logback/commit/f46044b805bca91efe5fd6afe52257cd02f775f8\", \"fixResolution\": \"logback-classic/src/main/java/ch/qos/logback/classic/net/server/LogbackClassicSerializationHelper.java,logback-core/src/test/java/ch/qos/logback/core/net/Innocent.java,logback-classic/src/main/java/ch/qos/logback/classic/net/SimpleSocketServer.java,logback-core/src/main/java/ch/qos/logback/core/net/HardenedObjectInputStream.java,logback-classic/src/test/java/ch/qos/logback/classic/LoggerSerializationTest.java,logback-core/src/test/java/ch/qos/logback/core/net/HardenedObjectInputStreamTest.java\", \"date\": \"2017-02-07\", \"message\": \"harden serialization\", \"extraData\": \"key=f46044b&committerName=ceki&committerUrl=https://github.com/ceki&committerAvatar=https://avatars1.githubusercontent.com/u/115476?v=4\" }," +
                "{ \"vulnerability\": \"CVE-2017-5929\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/victims/victims-cve-db/commit/94745e07d4b10a58d98ffc3916d9711fa407b018\", \"fixResolution\": \"database/java/2017/5929.yaml\", \"date\": \"2017-03-15\", \"message\": \"Added CVE-2017-5929 per issue #76\", \"extraData\": \"key=94745e0&committerName=cplvic&committerUrl=https://github.com/cplvic&committerAvatar=https://avatars0.githubusercontent.com/u/11528385?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: LogbackClassicSerializationHelper.java, Innocent.java, SimpleSocketServer.java, HardenedObjectInputStream.java, LoggerSerializationTest.java, HardenedObjectInputStreamTest.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"d2b236f1-ac7d-4413-92fa-848cc7c3d3e5\", \"keyId\": 754708, \"filename\": \"logback-core-1.1.3.jar\", \"name\": \"Logback Core Module\", \"groupId\": \"ch.qos.logback\", \"artifactId\": \"logback-core\", \"version\": \"1.1.3\", \"sha1\": \"e3c02049f2dbbc764681b40094ecf0dcbc99b157\", \"type\": \"Java\"," +
                "\"references\": { \"url\": \"http://logback.qos.ch/logback-core\", \"pomUrl\": \"http://maven.ibiblio.org/maven2/ch/qos/logback/logback-core/1.1.3/logback-core-1.1.3.pom\", \"scmUrl\": \"https://github.com/ceki/logback/logback-core\" }, \"licenses\": [ { \"name\": \"Eclipse 1.0\", \"url\": \"http://opensource.org/licenses/eclipse-1.0.php\", \"profileInfo\": { \"copyrightRiskScore\": \"SIX\", \"patentRiskScore\": \"FOUR\", \"copyleft\": \"PARTIAL\", \"linking\": \"VIRAL\", \"royaltyFree\": \"CONDITIONAL\" } }, { \"name\": \"LGPL 2.1\", \"url\": \"http://opensource.org/licenses/lgpl-2.1\", \"profileInfo\": { \"copyrightRiskScore\": \"FIVE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"PARTIAL\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:1,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2012-4529\", \"type\": \"CVE\", \"severity\": \"medium\", \"score\": 4.3, \"cvss3_score\": 0," +
                "\"publishDate\": \"2013-10-28\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2012-4529\", \"description\": \"The org.apache.catalina.connector.Response.encodeURL method in Red Hat JBoss Web 7.1.x and earlier, when the tracking mode is set to COOKIE, sends the jsessionid in the URL of the first response of a session, which allows remote attackers to obtain the session id (1) via a man-in-the-middle attack or (2) by reading a log.\", \"allFixes\": [], \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"aa3a10d7-4e2c-46fe-bbf9-3c2d06e43b02\", \"keyId\": 24276785, \"filename\": \"tomcat-embed-core-7.0.78.jar\", \"name\": \"tomcat-embed-core\", \"groupId\": \"org.apache.tomcat.embed\", \"artifactId\": \"tomcat-embed-core\", \"version\": \"7.0.78\", \"sha1\": \"ddb63d615ec3944b4394aed6dc825cd0cbb16b21\", \"type\": \"Java\", \"references\": { \"url\": \"http://tomcat.apache.org/\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/org/apache/tomcat/embed/tomcat-embed-core/7.0.78/tomcat-embed-core-7.0.78.pom\" }, \"licenses\": [ { \"name\": \"Apache 2.0\"," +
                "\"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": true, \"description\": \"High:3,Medium:2,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2017-12615\", \"type\": \"CVE\", \"severity\": \"medium\", \"score\": 6.8, \"cvss3_severity\": \"high\", \"cvss3_score\": 8.1, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2017-09-19\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-12615\", \"description\": \"When running Apache Tomcat 7.0.0 to 7.0.79 on Windows with HTTP PUTs enabled (e.g. via setting the readonly initialisation parameter of the Default to false) it was possible to upload a JSP file to the server via a specially crafted request. This JSP could then be requested and any code it contained would be executed by the server.\"," +
                "\"topFix\": { \"vulnerability\": \"CVE-2017-12615\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"SECURITY_TRACKER\", \"url\": \"http://www.securitytracker.com/id/1039392\", \"fixResolution\": \"The vendor has issued a fix (7.0.81).\\n\\nThe vendor advisory is available at:\\n\\nhttps://tomcat.apache.org/security-7.html#Fixed_in_Apache_Tomcat_7.0.81\", \"date\": \"2017-12-31\", \"message\": \"Apache Tomcat on Windows HTTP PUT Request Processing Flaw Lets Remote Users Execute Arbitrary Code on the Target System\", \"extraData\": \"key=1039392\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-12615\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"SECURITY_TRACKER\", \"url\": \"http://www.securitytracker.com/id/1039392\", \"fixResolution\": \"The vendor has issued a fix (7.0.81).\\n\\nThe vendor advisory is available at:\\n\\nhttps://tomcat.apache.org/security-7.html#Fixed_in_Apache_Tomcat_7.0.81\", \"date\": \"2017-12-31\", \"message\": \"Apache Tomcat on Windows HTTP PUT Request Processing Flaw Lets Remote Users Execute Arbitrary Code on the Target System\", \"extraData\": \"key=1039392\" }, { \"vulnerability\": \"CVE-2017-12615\", \"type\": \"CHANGE_FILES\"," +
                "\"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat70/commit/07dc0ea2745f0afab6415f22b16a29f1c6de5727\", \"fixResolution\": \"java/org/apache/naming/resources/VirtualDirContext.java,webapps/docs/changelog.xml,java/org/apache/naming/resources/FileDirContext.java\", \"date\": \"2017-08-10\", \"message\": \"Correct regression in r1804604 that broke WebDAV.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc7.0.x/trunk@1804729 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=07dc0ea&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" } ], \"fixResolutionText\": \"The vendor has issued a fix (7.0.81).\\n\\nThe vendor advisory is available at:\\n\\nhttps://tomcat.apache.org/security-7.html#Fixed_in_Apache_Tomcat_7.0.81\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"aa3a10d7-4e2c-46fe-bbf9-3c2d06e43b02\", \"keyId\": 24276785, \"filename\": \"tomcat-embed-core-7.0.78.jar\", \"name\": \"tomcat-embed-core\", \"groupId\": \"org.apache.tomcat.embed\", \"artifactId\": \"tomcat-embed-core\"," +
                "\"version\": \"7.0.78\", \"sha1\": \"ddb63d615ec3944b4394aed6dc825cd0cbb16b21\", \"type\": \"Java\", \"references\": { \"url\": \"http://tomcat.apache.org/\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/org/apache/tomcat/embed/tomcat-embed-core/7.0.78/tomcat-embed-core-7.0.78.pom\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": true, \"description\": \"High:3,Medium:2,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2017-12616\", \"type\": \"CVE\", \"severity\": \"medium\", \"score\": 5, \"cvss3_severity\": \"high\", \"cvss3_score\": 7.5, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N\", \"publishDate\": \"2017-09-19\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-12616\"," +
                "\"description\": \"When using a VirtualDirContext with Apache Tomcat 7.0.0 to 7.0.80 it was possible to bypass security constraints and/or view the source code of JSPs for resources served by the VirtualDirContext using a specially crafted request.\", \"topFix\": { \"vulnerability\": \"CVE-2017-12616\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"SECURITY_TRACKER\", \"url\": \"http://www.securitytracker.com/id/1039393\", \"fixResolution\": \"The vendor has issued a fix (7.0.81).\\n\\nThe vendor advisory is available at:\\n\\nhttps://tomcat.apache.org/security-7.html#Fixed_in_Apache_Tomcat_7.0.81\", \"date\": \"2017-12-31\", \"message\": \"Apache Tomcat VirtualDirContext Flaw Lets Remote Users View JSP Source Code for the Affected Resource\", \"extraData\": \"key=1039393\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-12616\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"SECURITY_TRACKER\", \"url\": \"http://www.securitytracker.com/id/1039393\", \"fixResolution\": \"The vendor has issued a fix (7.0.81).\\n\\nThe vendor advisory is available at:\\n\\nhttps://tomcat.apache.org/security-7.html#Fixed_in_Apache_Tomcat_7.0.81\", \"date\": \"2017-12-31\"," +
                "\"message\": \"Apache Tomcat VirtualDirContext Flaw Lets Remote Users View JSP Source Code for the Affected Resource\", \"extraData\": \"key=1039393\" }, { \"vulnerability\": \"CVE-2017-12616\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat70/commit/07dc0ea2745f0afab6415f22b16a29f1c6de5727\", \"fixResolution\": \"java/org/apache/naming/resources/VirtualDirContext.java,webapps/docs/changelog.xml,java/org/apache/naming/resources/FileDirContext.java\", \"date\": \"2017-08-10\", \"message\": \"Correct regression in r1804604 that broke WebDAV.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc7.0.x/trunk@1804729 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=07dc0ea&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" } ], \"fixResolutionText\": \"The vendor has issued a fix (7.0.81).\\n\\nThe vendor advisory is available at:\\n\\nhttps://tomcat.apache.org/security-7.html#Fixed_in_Apache_Tomcat_7.0.81\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\"," +
                "\"library\": { \"keyUuid\": \"aa3a10d7-4e2c-46fe-bbf9-3c2d06e43b02\", \"keyId\": 24276785, \"filename\": \"tomcat-embed-core-7.0.78.jar\", \"name\": \"tomcat-embed-core\", \"groupId\": \"org.apache.tomcat.embed\", \"artifactId\": \"tomcat-embed-core\", \"version\": \"7.0.78\", \"sha1\": \"ddb63d615ec3944b4394aed6dc825cd0cbb16b21\", \"type\": \"Java\", \"references\": { \"url\": \"http://tomcat.apache.org/\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/org/apache/tomcat/embed/tomcat-embed-core/7.0.78/tomcat-embed-core-7.0.78.pom\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": true, \"description\": \"High:3,Medium:2,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2017-7674\", \"type\": \"CVE\", \"severity\": \"medium\", \"score\": 4.3," +
                "\"cvss3_severity\": \"medium\", \"cvss3_score\": 4.3, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:R/S:U/C:N/I:L/A:N\", \"publishDate\": \"2017-08-11\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-7674\", \"description\": \"The CORS Filter in Apache Tomcat 9.0.0.M1 to 9.0.0.M21, 8.5.0 to 8.5.15, 8.0.0.RC1 to 8.0.44 and 7.0.41 to 7.0.78 did not add an HTTP Vary header indicating that the response varies depending on Origin. This permitted client and server side cache poisoning in some circumstances.\", \"topFix\": { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"BUGZILLA\", \"url\": \"https://bugzilla.redhat.com/show_bug.cgi?id=CVE-2017-7674\", \"fixResolution\": \"tomcat 7.0.79,tomcat 8.0.45,tomcat 8.5.16\", \"message\": \"CVE-2017-7674 tomcat: Vary header not added by CORS filter leading to cache poisoning\", \"extraData\": \"key=1480618&assignee=Red Hat Product Security\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"BUGZILLA\", \"url\": \"https://bugzilla.redhat.com/show_bug.cgi?id=CVE-2017-7674\"," +
                "\"fixResolution\": \"tomcat 7.0.79,tomcat 8.0.45,tomcat 8.5.16\", \"message\": \"CVE-2017-7674 tomcat: Vary header not added by CORS filter leading to cache poisoning\", \"extraData\": \"key=1480618&assignee=Red Hat Product Security\" }, { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"UPGRADE_VERSION\", \"origin\": \"BUGZILLA\", \"url\": \"https://bugzilla.redhat.com/show_bug.cgi?id=1480618\", \"fixResolution\": \"tomcat 7.0.79,tomcat 8.0.45,tomcat 8.5.16\", \"message\": \"CVE-2017-7674 tomcat: Vary header not added by CORS filter leading to cache poisoning\", \"extraData\": \"key=1480618&assignee=Red Hat Product Security\" }, { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat/commit/b94478d45b7e1fc06134a785571f78772fa30fed\", \"fixResolution\": \"java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml\", \"date\": \"2017-05-22\", \"message\": \"BZ61101: CORS filter should set Vary header in response. Submitted by Rick Riemer.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/trunk@1795813 13f79535-47bb-0310-9956-ffa450edef68\"," +
                "\"extraData\": \"key=b94478d&committerName=rmaucher&committerUrl=https://github.com/rmaucher&committerAvatar=https://avatars2.githubusercontent.com/u/324250?v=4\" }, { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat85/commit/9044c1672bbe4b2cf4c55028cc8b977cc62650e7\", \"fixResolution\": \"java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml\", \"date\": \"2017-05-22\", \"message\": \"BZ61101: CORS filter should set Vary header in response. Submitted by Rick Riemer.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc8.5.x/trunk@1795814 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=9044c16&committerName=rmaucher&committerUrl=https://github.com/rmaucher&committerAvatar=https://avatars2.githubusercontent.com/u/324250?v=4\" }, { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat70/commit/52382ebfbce20a98b01cd9d37184a12703987a5a\", \"fixResolution\": \"java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml\"," +
                "\"date\": \"2017-05-22\", \"message\": \"BZ61101: CORS filter should set Vary header in response. Submitted by Rick Riemer.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc7.0.x/trunk@1795816 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=52382eb&committerName=rmaucher&committerUrl=https://github.com/rmaucher&committerAvatar=https://avatars2.githubusercontent.com/u/324250?v=4\" }, { \"vulnerability\": \"CVE-2017-7674\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat80/commit/f52c242d92d4563dd1226dcc993ec37370ba9ce3\", \"fixResolution\": \"java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml\", \"date\": \"2017-05-22\", \"message\": \"BZ61101: CORS filter should set Vary header in response. Submitted by Rick Riemer.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc8.0.x/trunk@1795815 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=f52c242&committerName=rmaucher&committerUrl=https://github.com/rmaucher&committerAvatar=https://avatars2.githubusercontent.com/u/324250?v=4\" } ], \"fixResolutionText\": \"Upgrade to version tomcat 7.0.79, tomcat 8.0.45, tomcat 8.5.16 or greater\"," +
                "\"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"aa3a10d7-4e2c-46fe-bbf9-3c2d06e43b02\", \"keyId\": 24276785, \"filename\": \"tomcat-embed-core-7.0.78.jar\", \"name\": \"tomcat-embed-core\", \"groupId\": \"org.apache.tomcat.embed\", \"artifactId\": \"tomcat-embed-core\", \"version\": \"7.0.78\", \"sha1\": \"ddb63d615ec3944b4394aed6dc825cd0cbb16b21\", \"type\": \"Java\", \"references\": { \"url\": \"http://tomcat.apache.org/\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/org/apache/tomcat/embed/tomcat-embed-core/7.0.78/tomcat-embed-core-7.0.78.pom\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": true, \"description\": \"High:3,Medium:2,\", \"date\": \"2017-10-26\" }, { \"vulnerability\": { \"name\": \"CVE-2018-8014\"," +
                "\"type\": \"CVE\", \"severity\": \"high\", \"score\": 7.5, \"cvss3_severity\": \"high\", \"cvss3_score\": 9.8, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-05-16\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-8014\", \"description\": \"The defaults settings for the CORS filter provided in Apache Tomcat 9.0.0.M1 to 9.0.8, 8.5.0 to 8.5.31, 8.0.0.RC1 to 8.0.52, 7.0.41 to 7.0.88 are insecure and enable 'supportsCredentials' for all origins. It is expected that users of the CORS filter will have configured it appropriately for their environment rather than using it in the default configuration. Therefore, it is expected that most users will not be impacted by this issue.\", \"topFix\": { \"vulnerability\": \"CVE-2018-8014\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat70/commit/5877390a9605f56d9bd6859a54ccbfb16374a78b\", \"fixResolution\": \"java/org/apache/catalina/filters/LocalStrings.properties,test/org/apache/catalina/filters/TestCorsFilter.java,java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml,test/org/apache/catalina/filters/TesterFilterConfigs.java\"," +
                "\"date\": \"2018-05-16\", \"message\": \"Fix https://bz.apache.org/bugzilla/show_bug.cgi?id=62343\\nMake CORS filter defaults more secure.\\nThis is the fix for CVE-2018-8014.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc7.0.x/trunk@1831730 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=5877390&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2018-8014\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat70/commit/5877390a9605f56d9bd6859a54ccbfb16374a78b\", \"fixResolution\": \"java/org/apache/catalina/filters/LocalStrings.properties,test/org/apache/catalina/filters/TestCorsFilter.java,java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml,test/org/apache/catalina/filters/TesterFilterConfigs.java\", \"date\": \"2018-05-16\", \"message\": \"Fix https://bz.apache.org/bugzilla/show_bug.cgi?id=62343\\nMake CORS filter defaults more secure.\\nThis is the fix for CVE-2018-8014.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc7.0.x/trunk@1831730 13f79535-47bb-0310-9956-ffa450edef68\"," +
                "\"extraData\": \"key=5877390&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" }, { \"vulnerability\": \"CVE-2018-8014\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat80/commit/2c9d8433bd3247a2856d4b2555447108758e813e#diff-32f241c95d21b1b224601e52f83af334\", \"fixResolution\": \"java/org/apache/catalina/filters/LocalStrings.properties,test/org/apache/catalina/filters/TestCorsFilter.java,java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml,test/org/apache/catalina/filters/TesterFilterConfigs.java\", \"date\": \"2018-05-16\", \"message\": \"Fix https://bz.apache.org/bugzilla/show_bug.cgi?id=62343\\nMake CORS filter defaults more secure.\\nThis is the fix for CVE-2018-8014.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc8.0.x/trunk@1831729 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=2c9d843&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" }, { \"vulnerability\": \"CVE-2018-8014\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\"," +
                "\"url\": \"https://github.com/apache/tomcat/commit/d83a76732e6804739b81d8b2056365307637b42d\", \"fixResolution\": \"java/org/apache/catalina/filters/LocalStrings.properties,test/org/apache/catalina/filters/TestCorsFilter.java,java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml,test/org/apache/catalina/filters/TesterFilterConfigs.java\", \"date\": \"2018-05-16\", \"message\": \"Fix https://bz.apache.org/bugzilla/show_bug.cgi?id=62343\\nMake CORS filter defaults more secure.\\nThis is the fix for CVE-2018-8014.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/trunk@1831726 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=d83a767&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" }, { \"vulnerability\": \"CVE-2018-8014\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/apache/tomcat85/commit/60f596a21fd6041335a3a1a4015d4512439cecb5\", \"fixResolution\": \"java/org/apache/catalina/filters/LocalStrings.properties,test/org/apache/catalina/filters/TestCorsFilter.java,java/org/apache/catalina/filters/CorsFilter.java,webapps/docs/changelog.xml,test/org/apache/catalina/filters/TesterFilterConfigs.java\"," +
                "\"date\": \"2018-05-16\", \"message\": \"Fix https://bz.apache.org/bugzilla/show_bug.cgi?id=62343\\nMake CORS filter defaults more secure.\\nThis is the fix for CVE-2018-8014.\\n\\ngit-svn-id: https://svn.apache.org/repos/asf/tomcat/tc8.5.x/trunk@1831728 13f79535-47bb-0310-9956-ffa450edef68\", \"extraData\": \"key=60f596a&committerName=markt-asf&committerUrl=https://github.com/markt-asf&committerAvatar=https://avatars3.githubusercontent.com/u/4690029?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: LocalStrings.properties, TestCorsFilter.java, CorsFilter.java, changelog.xml, TesterFilterConfigs.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"aa3a10d7-4e2c-46fe-bbf9-3c2d06e43b02\", \"keyId\": 24276785, \"filename\": \"tomcat-embed-core-7.0.78.jar\", \"name\": \"tomcat-embed-core\", \"groupId\": \"org.apache.tomcat.embed\", \"artifactId\": \"tomcat-embed-core\", \"version\": \"7.0.78\", \"sha1\": \"ddb63d615ec3944b4394aed6dc825cd0cbb16b21\", \"type\": \"Java\", \"references\": { \"url\": \"http://tomcat.apache.org/\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/org/apache/tomcat/embed/tomcat-embed-core/7.0.78/tomcat-embed-core-7.0.78.pom\" }," +
                "\"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": true,            \"description\": \"High:3,Medium:2,\", \"date\": \"2017-10-26\" } ] }").alerts
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
                "projectVersion": "1.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [ { \"vulnerability\": { \"name\": \"CVE-2017-15095\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 2.1, \"cvss3_severity\": \"high\", \"cvss3_score\": 5.3, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-06\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-15095\", \"description\": \"A deserialization flaw was discovered in the jackson-databind in versions before 2.8.10 and 2.9.1, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper. This issue extends the previous flaw CVE-2017-7525 by blacklisting more classes that could be used maliciously.\", \"topFix\": { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\"," +
                "\"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\", \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\"," +
                "\"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e865a7a4464da63ded9f4b1a2328ad85c9ded78b#diff-98084d808198119d550a9211e128a16f\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-12-12\", \"message\": \"Fix #1737 (#1857)\", \"extraData\": \"key=e865a7a&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\"," +
                "\"url\": \"https://github.com/FasterXML/jackson-databind/commit/e8f043d1\", \"fixResolution\": \"release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-06-30\", \"message\": \"Fix #1680\", \"extraData\": \"key=e8f043d&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: IllegalTypesCheckTest.java, VERSION, BeanDeserializerFactory.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\"," +
                "\"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" } ] }").alerts
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
                "projectVersion": "1.0.0"
            ])
            return result
        })
        helper.registerAllowedMethod("fetchVulnerabilities", [List], {
            return new JsonUtils().jsonStringToGroovyObject("{ \"alerts\": [ { \"vulnerability\": { \"name\": \"CVE-2017-15095\", \"type\": \"CVE\", \"severity\": \"high\", \"score\": 2.1, \"cvss3_severity\": \"high\", \"cvss3_score\": 5.3, \"scoreMetadataVector\": \"CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H\", \"publishDate\": \"2018-02-06\", \"url\": \"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2017-15095\", \"description\": \"A deserialization flaw was discovered in the jackson-databind in versions before 2.8.10 and 2.9.1, which could allow an unauthenticated user to perform code execution by sending the maliciously crafted input to the readValue method of the ObjectMapper. This issue extends the previous flaw CVE-2017-7525 by blacklisting more classes that could be used maliciously.\", \"topFix\": { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\"," +
                "\"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\", \"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, \"allFixes\": [ { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/60d459ce\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-04-13\", \"message\": \"Fix #1599 for 2.8.9\\n\\nMerge branch '2.7' into 2.8\"," +
                "\"extraData\": \"key=60d459c&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\", \"url\": \"https://github.com/FasterXML/jackson-databind/commit/e865a7a4464da63ded9f4b1a2328ad85c9ded78b#diff-98084d808198119d550a9211e128a16f\", \"fixResolution\": \"src/test/java/com/fasterxml/jackson/databind/interop/IllegalTypesCheckTest.java,release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-12-12\", \"message\": \"Fix #1737 (#1857)\", \"extraData\": \"key=e865a7a&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" }, { \"vulnerability\": \"CVE-2017-15095\", \"type\": \"CHANGE_FILES\", \"origin\": \"GITHUB_COMMIT\"," +
                "\"url\": \"https://github.com/FasterXML/jackson-databind/commit/e8f043d1\", \"fixResolution\": \"release-notes/VERSION,src/main/java/com/fasterxml/jackson/databind/deser/BeanDeserializerFactory.java\", \"date\": \"2017-06-30\", \"message\": \"Fix #1680\", \"extraData\": \"key=e8f043d&committerName=cowtowncoder&committerUrl=https://github.com/cowtowncoder&committerAvatar=https://avatars0.githubusercontent.com/u/55065?v=4\" } ], \"fixResolutionText\": \"Replace or update the following files: IllegalTypesCheckTest.java, VERSION, BeanDeserializerFactory.java\", \"references\": [] }, \"type\": \"SECURITY_VULNERABILITY\", \"level\": \"MAJOR\", \"library\": { \"keyUuid\": \"13f7802e-8aa1-4303-a5db-1d0c85e871a9\", \"keyId\": 23410061, \"filename\": \"jackson-databind-2.8.8.jar\", \"name\": \"jackson-databind\", \"groupId\": \"com.fasterxml.jackson.core\", \"artifactId\": \"jackson-databind\", \"version\": \"2.8.8\", \"sha1\": \"bf88c7b27e95cbadce4e7c316a56c3efffda8026\"," +
                "\"type\": \"Java\", \"references\": { \"url\": \"http://github.com/FasterXML/jackson\", \"issueUrl\": \"https://github.com/FasterXML/jackson-databind/issues\", \"pomUrl\": \"http://repo.jfrog.org/artifactory/list/repo1/com/fasterxml/jackson/core/jackson-databind/2.8.8/jackson-databind-2.8.8.pom\", \"scmUrl\": \"http://github.com/FasterXML/jackson-databind\" }, \"licenses\": [ { \"name\": \"Apache 2.0\", \"url\": \"http://apache.org/licenses/LICENSE-2.0\", \"profileInfo\": { \"copyrightRiskScore\": \"THREE\", \"patentRiskScore\": \"ONE\", \"copyleft\": \"NO\", \"linking\": \"DYNAMIC\", \"royaltyFree\": \"CONDITIONAL\" } } ] }, \"project\": \"pipeline-test - 0.0.1\", \"projectId\": 302194, \"projectToken\": \"1b8fdc36cb6949f482d0fd936a39dab69d6b34f43fff4dda8a9241f2c6e536c7\", \"directDependency\": false, \"description\": \"High:5,\", \"date\": \"2017-11-15\" } ] }").alerts
        })

        try {
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
        } catch (e) {
            assertThat(e.getMessage(), containsString('[whitesourceExecuteScan] 1 Open Source Software Security vulnerabilities with CVSS score greater or equal 0 detected. - '))
            assertThat(writeFileRule.files['piper_whitesource_vulnerability_report.json'], not(isEmptyOrNullString()))
        }
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
                "projectVersion": "1.0.0"
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
        def error = false
        try {
            stepRule.step.checkStatus(0, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
        }
        assertThat(error, is(false))
    }

    @Test
    void testCheckStatus_255() {
        def error = false
        try {
            stepRule.step.checkStatus(255, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] The scan resulted in an error"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_254() {
        def error = false
        try {
            stepRule.step.checkStatus(254, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] Whitesource found one or multiple policy violations"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_253() {
        def error = false
        try {
            stepRule.step.checkStatus(253, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] The local scan client failed to execute the scan"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_252() {
        def error = false
        try {
            stepRule.step.checkStatus(252, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] There was a failure in the connection to the WhiteSource servers"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_251() {
        def error = false
        try {
            stepRule.step.checkStatus(251, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] The server failed to analyze the scan"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_250() {
        def error = false
        try {
            stepRule.step.checkStatus(250, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] Pre-step failure"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_127() {
        def error = false
        try {
            stepRule.step.checkStatus(127, [whitesource:[licensingVulnerabilities: true]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] Whitesource scan failed with unknown error code '127'"))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckStatus_vulnerability() {
        def error = false
        try {
            stepRule.step.checkStatus(0, [whitesource:[licensingVulnerabilities: false, securityVulnerabilities: true, severeVulnerabilities: 5, cvssSeverityLimit: 7]])
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] 5 Open Source Software Security vulnerabilities with CVSS score greater or equal 7 detected. - "))
        }
        assertThat(error, is(true))
    }

    @Test
    void testCheckViolationStatus_0() {
        def error = false
        try {
            stepRule.step.checkViolationStatus(0)
        } catch (e) {
            error = true
        }
        assertThat(error, is(false))
        assertThat(loggingRule.log, containsString("[whitesourceExecuteScan] No policy violations found"))
    }

    @Test
    void testCheckViolationStatus_5() {
        def error = false
        try {
            stepRule.step.checkViolationStatus(5)
        } catch (e) {
            error = true
            assertThat(e.getMessage(), is("[whitesourceExecuteScan] Whitesource found 5 policy violations for your product"))
        }
        assertThat(error, is(true))
    }
}
