package com.sap.piper

import org.junit.Assert
import org.junit.Before
import org.junit.Test

class ConfigurationLoaderTest {

    @Before
    public void init() {
        Map configuration = [:]
        configuration.general = [productiveBranch: 'master']
        configuration.steps = [executeMaven: [dockerImage: 'maven:3.2-jdk-8-onbuild']]
        configuration.stages = [staticCodeChecks: [pmdExcludes: '**']]
        configuration.postActions = [sendEmail: [recipients: 'myEmail']]

        Map defaultConfiguration = [:]
        defaultConfiguration.general = [productiveBranch: 'develop']
        defaultConfiguration.steps = [executeGradle: [dockerImage: 'gradle:4.0.1-jdk8']]
        defaultConfiguration.stages = [staticCodeChecks: [pmdExcludes: '*.java']]

        DefaultValueCache.createInstance(defaultConfiguration, configuration)
    }

    @Test
    void testLoadStepConfiguration() {
        Map config = ConfigurationLoader.stepConfiguration('executeMaven')
        Assert.assertEquals('maven:3.2-jdk-8-onbuild', config.dockerImage)
    }

    @Test
    void testLoadStageConfiguration() {
        Map config = ConfigurationLoader.stageConfiguration('staticCodeChecks')
        Assert.assertEquals('**', config.pmdExcludes)
    }

    @Test
    void testLoadGeneralConfiguration() {
        Map config = ConfigurationLoader.generalConfiguration()
        Assert.assertEquals('master', config.productiveBranch)
    }

    @Test
    void testLoadDefaultStepConfiguration() {
        Map config = ConfigurationLoader.defaultStepConfiguration('executeGradle')
        Assert.assertEquals('gradle:4.0.1-jdk8', config.dockerImage)
    }

    @Test
    void testLoadDefaultStageConfiguration() {
        Map config = ConfigurationLoader.defaultStageConfiguration('staticCodeChecks')
        Assert.assertEquals('*.java', config.pmdExcludes)
    }

    @Test
    void testLoadDefaultGeneralConfiguration() {
        Map config = ConfigurationLoader.defaultGeneralConfiguration()
        Assert.assertEquals('develop', config.productiveBranch)
    }

    @Test
    void testLoadPostActionConfiguration(){
        Map config = ConfigurationLoader.postActionConfiguration('sendEmail')
        Assert.assertEquals('myEmail', config.recipients)
    }
}
