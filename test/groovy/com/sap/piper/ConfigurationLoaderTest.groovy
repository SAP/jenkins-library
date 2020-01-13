package com.sap.piper

import org.junit.After
import org.junit.Assert
import org.junit.Before
import org.junit.Test

class ConfigurationLoaderTest {

    static def mockScript = new Script() {
        def commonPipelineEnvironment

        @Override
        Object run() {
            return null
        }
    }

    @Before
    void setUp() {
        Map configuration = [:]
        configuration.general = [productiveBranch: 'master']
        configuration.steps = [executeMaven: [dockerImage: 'maven:3.2-jdk-8-onbuild']]
        configuration.stages = [staticCodeChecks: [pmdExcludes: '**']]
        configuration.postActions = [sendEmail: [recipients: 'myEmail']]

        mockScript.commonPipelineEnvironment = [configuration: configuration]

        Map defaultConfiguration = [:]
        defaultConfiguration.general = [productiveBranch: 'develop']
        defaultConfiguration.steps = [executeGradle: [dockerImage: 'gradle:4.0.1-jdk8']]
        defaultConfiguration.stages = [staticCodeChecks: [pmdExcludes: '*.java']]

        DefaultValueCache.createInstance(defaultConfiguration, mockScript.getBinding())
    }

    @After
    void tearDown() {
        DefaultValueCache.reset()
    }

    @Test
    void testLoadStepConfiguration() {
        Map config = ConfigurationLoader.stepConfiguration(mockScript, 'executeMaven')
        Assert.assertEquals('maven:3.2-jdk-8-onbuild', config.dockerImage)
    }

    @Test
    void testLoadStageConfiguration() {
        Map config = ConfigurationLoader.stageConfiguration(mockScript, 'staticCodeChecks')
        Assert.assertEquals('**', config.pmdExcludes)
    }

    @Test
    void testLoadGeneralConfiguration() {
        Map config = ConfigurationLoader.generalConfiguration(mockScript)
        Assert.assertEquals('master', config.productiveBranch)
    }

    @Test
    void testLoadDefaultStepConfiguration() {
        Map config = ConfigurationLoader.defaultStepConfiguration(mockScript, 'executeGradle')
        Assert.assertEquals('gradle:4.0.1-jdk8', config.dockerImage)
    }

    @Test
    void testLoadDefaultStageConfiguration() {
        Map config = ConfigurationLoader.defaultStageConfiguration(mockScript, 'staticCodeChecks')
        Assert.assertEquals('*.java', config.pmdExcludes)
    }

    @Test
    void testLoadDefaultGeneralConfiguration() {
        Map config = ConfigurationLoader.defaultGeneralConfiguration(mockScript)
        Assert.assertEquals('develop', config.productiveBranch)
    }

    @Test
    void testLoadPostActionConfiguration() {
        Map config = ConfigurationLoader.postActionConfiguration(mockScript, 'sendEmail')
        Assert.assertEquals('myEmail', config.recipients)
    }
}
