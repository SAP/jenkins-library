package com.sap.piper

import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsErrorRule
import util.JenkinsLoggingRule
import util.JenkinsSetupRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.hamcrest.core.Is.*

class DescriptorUtilsTest extends BasePiperTest {

    @Rule
    public JenkinsErrorRule errorRule = new JenkinsErrorRule(this)
    @Rule
    public JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    @Rule
    public JenkinsSetupRule setUpRule = new JenkinsSetupRule(this)
    @Rule
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
            .around(loggingRule)

    DescriptorUtils descriptorUtils

    @Before
    void init() throws Exception {
        descriptorUtils = new DescriptorUtils()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(descriptorUtils)
    }

    @Test
    void testGetNpmGAVSapArtifact() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/npm/${searchConfig.file}")
                return new JsonUtils().jsonStringToGroovyObject(packageJsonFile.text)
        })

        def gav = descriptorUtils.getNpmGAV('package2.json')

        assertEquals(gav.group, '')
        assertEquals(gav.artifact, 'some-test')
        assertEquals(gav.version, '1.2.3')
    }

    @Test
    void testGetNpmGAV() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/npm/${searchConfig.file}")
                return new JsonUtils().jsonStringToGroovyObject(packageJsonFile.text)
        })

        def gav = descriptorUtils.getNpmGAV('package.json')

        assertEquals(gav.group, '@sap')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetNpmGAVSapArtifactError() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/npm/${searchConfig.file}")
                return new JsonUtils().jsonStringToGroovyObject(packageJsonFile.text)
        })

        def errorCaught = false
        try {
            descriptorUtils.getNpmGAV('package3.json')
        } catch (e) {
            errorCaught = true
            assertThat(e, isA(AbortException.class))
            assertThat(e.getMessage(), is("Unable to parse package name '@someerror'"))
        }
        assertThat(errorCaught, is(true))
    }

    @Test
    void testGetSbtGAV() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/sbt/${searchConfig.file}")
                return new JsonUtils().jsonStringToGroovyObject(packageJsonFile.text)
        })

        def gav = descriptorUtils.getSbtGAV('sbtDescriptor.json')

        assertEquals(gav.group, 'sap')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.packaging, 'test')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetDubGAV() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/dub/${searchConfig.file}")
                return new JsonUtils().jsonStringToGroovyObject(packageJsonFile.text)
        })

        def gav = descriptorUtils.getDubGAV('dub.json')

        assertEquals(gav.group, 'com.sap.dlang')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetPipGAV() {

        helper.registerAllowedMethod("readFile", [Map.class], {
            map ->
                def descriptorFile = new File("test/resources/utilsTest/${map.file}")
                return descriptorFile.text
        })

        def gav = descriptorUtils.getPipGAV('setup.py')

        assertEquals('', gav.group)
        assertEquals('py_connect', gav.artifact)
        assertEquals('1.0', gav.version)
    }

    @Test
    void testGetPipGAVFromVersionTxt() {

        helper.registerAllowedMethod("readFile", [Map.class], {
            map ->
                def descriptorFile = new File("test/resources/DescriptorUtils/pip/${map.file}")
                return descriptorFile.text
        })

        def gav = descriptorUtils.getPipGAV('setup.py')

        assertEquals('', gav.group)
        assertEquals('some-test', gav.artifact)
        assertEquals('1.0.0-SNAPSHOT', gav.version)
    }

    @Test
    void testGetMavenGAVComplete() {

        helper.registerAllowedMethod("readMavenPom", [Map.class], {
            searchConfig ->
                return new Object(){
                    def groupId = 'test.group', artifactId = 'test-artifact', version = '1.2.4', packaging = 'jar'
                }
        })

        def gav = descriptorUtils.getMavenGAV('pom.xml')

        assertEquals(gav.group, 'test.group')
        assertEquals(gav.artifact, 'test-artifact')
        assertEquals(gav.version, '1.2.4')
        assertEquals(gav.packaging, 'jar')
    }

    @Test
    void testGetMavenGAVPartial() {
        def parameters = []

        helper.registerAllowedMethod("readMavenPom", [Map.class], {
            searchConfig ->
                return new Object(){
                    def groupId = null, artifactId = null, version = null, packaging = 'jar'
                }
        })

        helper.registerAllowedMethod("sh", [Map.class], {
            mvnHelpCommand ->
                def scriptCommand = mvnHelpCommand['script']
                parameters.add(scriptCommand)
                if(scriptCommand.contains('project.groupId'))
                    return 'test.group'
                if(scriptCommand.contains('project.artifactId'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.version'))
                    return '1.2.4'
        })

        def gav = descriptorUtils.getMavenGAV('pom.xml')

        assertEquals(gav.group, 'test.group')
        assertEquals(gav.artifact, 'test-artifact')
        assertEquals(gav.version, '1.2.4')
        assertEquals(gav.packaging, 'jar')
    }

    @Test
    void testGetGoGAV() {

        helper.registerAllowedMethod("readFile", [Map.class], {
            map ->
                def path = 'test/resources/DescriptorUtils/go/' + map.file.substring(map.file.lastIndexOf('/') + 1, map.file.length())
                def descriptorFile = new File(path)
                if(descriptorFile.exists())
                    return descriptorFile.text
                else
                    return null
        })

        def gav = descriptorUtils.getGoGAV('./myProject/Gopkg.toml', new URI('https://github.com/test/golang'))

        assertEquals('', gav.group)
        assertEquals('github.com/test/golang.myProject', gav.artifact)
        assertEquals('1.2.3', gav.version)
    }
}
