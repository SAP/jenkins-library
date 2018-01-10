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
}
