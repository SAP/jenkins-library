package com.sap.piper

import org.junit.Assert
import org.junit.Test

class ConfigurationMergerTest {

    @Test
    void testMerge(){
        Map defaults = [dockerImage: 'mvn']
        Map parameters = [goals: 'install', flags: '']
        List parameterKeys = ['flags']
        Map configuration = [flags: '-B']
        List configurationKeys = ['flags']
        Map merged = ConfigurationMerger.merge(parameters, parameterKeys, configuration, configurationKeys, defaults)
        Assert.assertEquals('mvn', merged.dockerImage)
        Assert.assertNull(merged.goals)
        Assert.assertEquals('', merged.flags)
    }

    @Test
    void testMergeParameterWithDefault(){
        Map defaults = [nonErpDestinations: []]
        Map parameters = [nonErpDestinations: null]
        List parameterKeys = ['nonErpDestinations']
        Map merged = ConfigurationMerger.merge(parameters, parameterKeys, defaults)
        Assert.assertEquals([], merged.nonErpDestinations)
    }

    @Test
    void testMergeCustomPipelineValues(){
        Map defaults = [dockerImage: 'mvn']
        Map parameters = [goals: 'install', flags: '']
        List parameterKeys = ['flags']
        Map configuration = [flags: '-B']
        List configurationKeys = ['flags']
        Map pipelineDataMap = [artifactVersion: '1.2.3', flags: 'test']
        Map merged = ConfigurationMerger.mergeWithPipelineData(parameters, parameterKeys, pipelineDataMap, configuration, configurationKeys, defaults)
        Assert.assertEquals('', merged.flags)
        Assert.assertEquals('1.2.3', merged.artifactVersion)
    }

    @Test
    void testMergeDeepStructure(){
        Map defaults = [fruits: [apples: 1, oranges: 10, bananaaas: 0]]
        Map configuration = [fruits: [bananaaas: 50, cucumbers: 1000]]
        List configurationKeys = ['fruits']
        Map parameters = [fruits: [apples: 18], veggie: []]
        List parameterKeys = ['fruits']
        Map merged = ConfigurationMerger.merge(parameters, parameterKeys, configuration, configurationKeys, defaults)
        Assert.assertEquals(50, merged.fruits.bananaaas)
        Assert.assertEquals(18, merged.fruits.apples)
        Assert.assertEquals(10, merged.fruits.oranges)
        Assert.assertEquals(1000, merged.fruits.cucumbers)
        Assert.assertEquals(null, merged.veggie)
    }

    @Test
    void testMergeGlobalAndStepConfiguration() {
        Map parameters = [priority: 1]
        List parameterKeys = ['priority']
        Map stepConfiguration = [priority: 2]
        List stepConfigurationKeys = ['priority']
        Map globalConfiguration = [priority: 3]
        List globalConfigurationKeys = ['priority']
        Map stepDefaults = [priority: 4]
        Map globalDefaults = [priority: 5]

        Map merged = ConfigurationMerger.merge(
            parameters, parameterKeys,
            stepConfiguration, stepConfigurationKeys, stepDefaults,
            globalConfiguration, globalConfigurationKeys, globalDefaults)

        assert merged.priority == 1

        merged = ConfigurationMerger.merge(
            [:], [],
            stepConfiguration, stepConfigurationKeys, stepDefaults,
            globalConfiguration, globalConfigurationKeys, globalDefaults)

        assert merged.priority == 2

        merged = ConfigurationMerger.merge(
            [:], [],
            [:], stepConfigurationKeys, stepDefaults,
            globalConfiguration, globalConfigurationKeys, globalDefaults)

        assert merged.priority == 3

        merged = ConfigurationMerger.merge(
            [:], [],
            [:], stepConfigurationKeys, stepDefaults,
            [:], [], globalDefaults)

        assert merged.priority == 4

        merged = ConfigurationMerger.merge(
            [:], [],
            [:], [], [:],
            [:], [], globalDefaults)

        assert merged.priority == 5
    }
}
