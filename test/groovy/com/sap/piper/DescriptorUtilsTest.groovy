package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsSetupRule
import util.LibraryLoadingTestExecutionListener
import util.SharedLibraryCreator

import static org.junit.Assert.assertEquals

class DescriptorUtilsTest extends BasePiperTest {

    @Rule
    public JenkinsSetupRule setUpRule = new JenkinsSetupRule(this, SharedLibraryCreator.lazyLoadedLibrary)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain =
        RuleChain.outerRule(setUpRule)
            .around(loggingRule)

    DescriptorUtils descriptorUtils

    @Before
    void init() throws Exception {
        nullScript.commonPipelineEnvironment = new Object() {
            def reset() {}
        }
        descriptorUtils = new DescriptorUtils()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(descriptorUtils)
    }

    @Test
    void testGetNpmGAV() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/npm/${searchConfig.file}")
                return new JsonUtils().parseJsonSerializable(packageJsonFile.text)
        })

        def gav = descriptorUtils.getNpmGAV('package.json')

        assertEquals(gav.group, '@sap')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetSbtGAV() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/sbt/${searchConfig.file}")
                return new JsonUtils().parseJsonSerializable(packageJsonFile.text)
        })

        def gav = descriptorUtils.getSbtGAV('sbtDescriptor.json')

        assertEquals(gav.group, 'sap')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.packaging, 'test')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetDlangGAV() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/DescriptorUtils/dlang/${searchConfig.file}")
                return new JsonUtils().parseJsonSerializable(packageJsonFile.text)
        })

        def gav = descriptorUtils.getDlangGAV('dub.json')

        assertEquals(gav.group, 'com.sap.dlang')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetPipGAV() {

        helper.registerAllowedMethod("sh", [Map.class], {
            map ->
                def descriptorFile = new File("test/resources/utilsTest/${map.script.substring(4, map.script.size())}")
                return descriptorFile.text
        })

        def gav = descriptorUtils.getPipGAV('setup.py')

        assertEquals('', gav.group)
        assertEquals('py_connect', gav.artifact)
        assertEquals('1.0', gav.version)
    }

    @Test
    void testGetPipGAVFromVersionTxt() {

        helper.registerAllowedMethod("sh", [Map.class], {
            map ->
                def descriptorFile = new File("test/resources/DescriptorUtils/pip/${map.script.substring(4, map.script.size())}")
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
}
