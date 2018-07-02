package com.sap.piper

import groovy.test.GroovyAssert

import static org.hamcrest.Matchers.*

import org.junit.Assert
import org.junit.Test

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
        Map config = new ConfigurationHelper([property1: '27']).use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '27'))
    }

    @Test
    void testConfigurationLoaderWithCustomSettings() {
        Map config = new ConfigurationHelper([property1: '27'])
            .mixin([property1: '41'])
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '41'))
    }

    @Test
    void testConfigurationLoaderWithFilteredCustomSettings() {
        Set filter = ['property2']
        Map config = new ConfigurationHelper([property1: '27'])
            .mixin([property1: '41', property2: '28', property3: '29'], filter)
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '27'))
        Assert.assertThat(config, hasEntry('property2', '28'))
        Assert.assertThat(config, not(hasKey('property3')))
    }

    @Test
    void testConfigurationLoaderWithBooleanValue() {
        Map config = new ConfigurationHelper([property1: '27'])
            .mixin([property1: false])
            .mixin([property2: false])
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', false))
        Assert.assertThat(config, hasEntry('property2', false))
    }

    @Test
    void testConfigurationLoaderWithMixinDependent() {
        Map config = new ConfigurationHelper([
                type: 'maven',
                maven: [dockerImage: 'mavenImage', dockerWorkspace: 'mavenWorkspace'],
                npm: [dockerImage: 'npmImage', dockerWorkspace: 'npmWorkspace', executeDocker: true, executeDocker3: false],
                executeDocker1: true
            ])
            .mixin([dockerImage: 'anyImage', type: 'npm', type2: 'npm', type3: '', executeDocker: false, executeDocker1: false, executeDocker2: false])
            .dependingOn('type').mixin('dockerImage')
            // test with empty dependent value
            .dependingOn('type3').mixin('dockerWorkspace')
            // test with empty dependent key
            .dependingOn('type4').mixin('dockerWorkspace')
            // test with empty default dependent value
            .dependingOn('type2').mixin('dockerWorkspace')
            // test with boolean value
            .dependingOn('type').mixin('executeDocker')
            .dependingOn('type').mixin('executeDocker2')
            .dependingOn('type').mixin('executeDocker3')
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('dockerImage', 'anyImage'))
        Assert.assertThat(config, hasEntry('dockerWorkspace', 'npmWorkspace'))
        Assert.assertThat(config, hasEntry('executeDocker', false))
        Assert.assertThat(config, hasEntry('executeDocker1', false))
        Assert.assertThat(config, hasEntry('executeDocker2', false))
        Assert.assertThat(config, hasEntry('executeDocker3', false))
    }

    @Test
    void testHandleCompatibility() {
        def configuration = new ConfigurationHelper([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'])
            .handleCompatibility(null, [old1: 'new1', old2: 'new2'])
            .use()

        Assert.assertThat(configuration.size(), is(5))
        Assert.assertThat(configuration.new1, is('oldValue1'))
        Assert.assertThat(configuration.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityNewAvailable() {
        def configuration = new ConfigurationHelper([old1: 'oldValue1', new1: 'newValue1', test: 'testValue'])
            .handleCompatibility(null, [old1: 'new1'])
            .use()

        Assert.assertThat(configuration.size(), is(3))
        Assert.assertThat(configuration.new1, is('newValue1'))
    }

    @Test
    void testHandleCompatibilityOldNotSet() {
        def configuration = new ConfigurationHelper([old1: null, test: 'testValue'])
            .handleCompatibility(null, [old1: 'new1'])
            .use()

        Assert.assertThat(configuration.size(), is(2))
    }
}
