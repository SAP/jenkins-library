package com.sap.piper

import org.junit.Assert
import org.junit.Test

class ConfigurationMergerTest {

    @Test
    void testMerge(){
        Map defaults = [dockerImage: 'mvn']
        Map parameters = [goals: 'install', flags: '']
        Set parameterKeys = ['flags']
        Map configuration = [flags: '-B']
        Set configurationKeys = ['flags']
        Map merged = ConfigurationMerger.merge(parameters, parameterKeys, configuration, configurationKeys, defaults)
        Assert.assertEquals('mvn', merged.dockerImage)
        Assert.assertNull(merged.goals)
        Assert.assertEquals('', merged.flags)
    }

    @Test
    void testMergeParameterWithDefault(){
        Map defaults = [nonErpDestinations: []]
        Map parameters = [nonErpDestinations: null]
        Set parameterKeys = ['nonErpDestinations']
        Map merged = ConfigurationMerger.merge(parameters, parameterKeys, defaults)
        Assert.assertEquals([], merged.nonErpDestinations)
    }

    @Test
    void testMergeCustomPipelineValues(){
        Map defaults = [dockerImage: 'mvn']
        Map parameters = [goals: 'install', flags: '']
        Set parameterKeys = ['flags']
        Map configuration = [flags: '-B']
        Set configurationKeys = ['flags']
        Map pipelineDataMap = [artifactVersion: '1.2.3', flags: 'test']
        Map merged = ConfigurationMerger.mergeWithPipelineData(parameters, parameterKeys, pipelineDataMap, configuration, configurationKeys, defaults)
        Assert.assertEquals('', merged.flags)
        Assert.assertEquals('1.2.3', merged.artifactVersion)
    }
}
