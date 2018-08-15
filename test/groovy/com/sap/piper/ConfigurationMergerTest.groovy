package com.sap.piper

import org.junit.Assert
import org.junit.Rule
import org.junit.Test

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.JenkinsResetDefaultCacheRule

class ConfigurationMergerTest {

    @Rule
    public JenkinsResetDefaultCacheRule resetDefaultValueCacheRule = new JenkinsResetDefaultCacheRule()

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
    void testMergeDeepStructure(){
        Map defaults = [fruits: [apples: 1, oranges: 10, bananas: 0]]
        Map configuration = [fruits: [bananas: 50, cucumbers: 1000]]
        Set configurationKeys = ['fruits']
        Map parameters = [fruits: [apples: 18], veggie: []]
        Set parameterKeys = ['fruits']
        Map merged = ConfigurationMerger.merge(parameters, parameterKeys, configuration, configurationKeys, defaults)
        Assert.assertEquals(50, merged.fruits.bananas)
        Assert.assertEquals(18, merged.fruits.apples)
        Assert.assertEquals(10, merged.fruits.oranges)
        Assert.assertEquals(1000, merged.fruits.cucumbers)
        Assert.assertEquals(null, merged.veggie)
    }

    @Test
    void testMergeDeepStructureWithMissingDefaults(){
        Map defaults = [others:[apples: 18]]
        Map configuration = [fruits: [bananas: 50, cucumbers: 1000]]
        Set configurationKeys = ['fruits']
        Map merged = ConfigurationMerger.merge(configuration, configurationKeys, defaults)
        Assert.assertEquals(50, merged.fruits.bananas)
        Assert.assertEquals(18, merged.others.apples)
        Assert.assertEquals(1000, merged.fruits.cucumbers)
    }
}
