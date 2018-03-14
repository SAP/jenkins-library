package com.sap.piper

import groovy.test.GroovyAssert
import org.junit.Assert
import org.junit.Test

import static org.hamcrest.Matchers.*

class ConfigurationHelperTest {

    private static getConfiguration() {
        Map configuration = [dockerImage: 'maven:3.2-jdk-8-onbuild']
        return configuration
    }

    @Test
    void testGetProperty() {
        def configuration = new ConfigurationHelper(getConfiguration())
        Assert.assertEquals('maven:3.2-jdk-8-onbuild', configuration.getConfigProperty('dockerImage'))
        Assert.assertEquals('maven:3.2-jdk-8-onbuild', configuration.getConfigProperty('dockerImage', 'default'))
        Assert.assertEquals('default', configuration.getConfigProperty('something', 'default'))
        Assert.assertTrue(configuration.isPropertyDefined('dockerImage'))
        Assert.assertFalse(configuration.isPropertyDefined('something'))
    }

    @Test
    void testIsPropertyDefined() {
        def configuration = new ConfigurationHelper(getConfiguration())
        Assert.assertTrue(configuration.isPropertyDefined('dockerImage'))
        Assert.assertFalse(configuration.isPropertyDefined('something'))
    }

    @Test
    void testIsPropertyDefinedWithInteger() {
        def configuration = new ConfigurationHelper([dockerImage: 3])
        Assert.assertTrue(configuration.isPropertyDefined('dockerImage'))
    }

    @Test
    void testGetMandatoryProperty() {
        def configuration = new ConfigurationHelper(getConfiguration())
        Assert.assertEquals('maven:3.2-jdk-8-onbuild', configuration.getMandatoryProperty('dockerImage'))
        Assert.assertEquals('default', configuration.getMandatoryProperty('something', 'default'))

        GroovyAssert.shouldFail { configuration.getMandatoryProperty('something') }
    }

    @Test
    void testConfigurationLoaderWithDefaults() {
        Map config = new ConfigurationHelper([property1: "27"]).use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', true))
    }

    @Test
    void testConfigurationLoaderWithCustomSettings() {
        Map config = new ConfigurationHelper([property1: "27"])
            .mixin([property1: "41"]])
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', false))
    }

    @Test
    void testConfigurationLoaderWithFilteredCustomSettings() {
        Map config = new ConfigurationHelper([property1: "27"])
            .mixin([property1: "41"]], ['property2'])
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', "27"))
    }
}
