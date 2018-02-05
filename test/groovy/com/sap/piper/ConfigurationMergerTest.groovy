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
        Map configurationKeys = [fruits: [apples: null, oranges: null, bananaaas: null]]
        Map parameters = [fruits: [apples: 18]]
        Map parameterKeys = [fruits: [apples: null, oranges: null, bananaaas: null]]
        Map merged = ConfigurationMerger.mergeDeepStructure(parameters, parameterKeys, configuration, configurationKeys, defaults)
        Assert.assertEquals(50, merged.fruits.bananaaas)
        Assert.assertEquals(18, merged.fruits.apples)
        Assert.assertEquals(10, merged.fruits.oranges)
        Assert.assertEquals(null, merged.fruits.cucumbers)
    }
}
