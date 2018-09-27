package com.sap.piper

import groovy.test.GroovyAssert

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

class ConfigurationHelperTest {

    @Rule
    public ExpectedException thrown = ExpectedException.none()

    private static getConfiguration() {
        Map configuration = [dockerImage: 'maven:3.2-jdk-8-onbuild']
        return configuration
    }

    @Test
    void testGetPropertyNestedLeafNodeIsString() {
        assertThat(ConfigurationHelper.getConfigPropertyNested([a:[b: 'c']], 'a/b'), is('c'))
    }

    @Test
    void testGetPropertyNestedLeafNodeIsMap() {
        assertThat(ConfigurationHelper.getConfigPropertyNested([a:[b: [c: 'd']]], 'a/b'), is([c: 'd']))
    }

    @Test
    void testGetPropertyNestedPathNotFound() {
        assertThat(ConfigurationHelper.getConfigPropertyNested([a:[b: 'c']], 'a/c'), is((nullValue())))
    }

    void testGetPropertyNestedPathStartsWithTokenizer() {
        assertThat(ConfigurationHelper.getConfigPropertyNested([k:'v'], '/k'), is(('v')))
    }

    @Test
    void testGetPropertyNestedPathEndsWithTokenizer() {
        assertThat(ConfigurationHelper.getConfigPropertyNested([k:'v'], 'k/'), is(('v')))
    }

    @Test
    void testGetPropertyNestedPathManyTokenizer() {
        assertThat(ConfigurationHelper.getConfigPropertyNested([k1:[k2 : 'v']], '///k1/////k2///'), is(('v')))
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
        def configuration = new ConfigurationHelper()
            .mixin([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'], null, null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(4))
        Assert.assertThat(configuration.newStructure.new1, is('oldValue1'))
        Assert.assertThat(configuration.newStructure.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityFlat() {
        def configuration = new ConfigurationHelper()
            .mixin([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'], null, null, [new1: 'old1', new2: 'old2'])
            .use()

        Assert.assertThat(configuration.size(), is(5))
        Assert.assertThat(configuration.new1, is('oldValue1'))
        Assert.assertThat(configuration.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityDeep() {
        def configuration = new ConfigurationHelper()
            .mixin([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'], null, null, [deep:[deeper:[newStructure: [new1: 'old1', new2: 'old2']]]])
            .use()

        Assert.assertThat(configuration.size(), is(4))
        Assert.assertThat(configuration.deep.deeper.newStructure.new1, is('oldValue1'))
        Assert.assertThat(configuration.deep.deeper.newStructure.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityNewAvailable() {
        def configuration = new ConfigurationHelper([old1: 'oldValue1', newStructure: [new1: 'newValue1'], test: 'testValue'])
            .mixin([old1: 'oldValue1', newStructure: [new1: 'newValue1'], test: 'testValue'], null, null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(3))
        Assert.assertThat(configuration.newStructure.new1, is('newValue1'))
    }

    @Test
    void testHandleCompatibilityOldNotSet() {
        def configuration = new ConfigurationHelper([old1: null, test: 'testValue'])
            .mixin([old1: null, test: 'testValue'], null, null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(2))
        Assert.assertThat(configuration.newStructure.new1, is(null))
    }

    @Test
    void testHandleCompatibilityNoneAvailable() {
        def configuration = new ConfigurationHelper([old1: null, test: 'testValue'])
            .mixin([test: 'testValue'], null, null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(2))
        Assert.assertThat(configuration.newStructure.new1, is(null))
    }

    @Test
    public void testWithMandoryParameterReturnDefaultFailureMessage() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR myKey')

        new ConfigurationHelper().withMandatoryProperty('myKey')
    }

    @Test
    public void testWithMandoryParameterReturnCustomerFailureMessage() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('My error message')

        new ConfigurationHelper().withMandatoryProperty('myKey', 'My error message')
    }

    @Test
    public void testWithMandoryParameterDefaultCustomFailureMessageProvidedSucceeds() {
        new ConfigurationHelper([myKey: 'myValue']).withMandatoryProperty('myKey', 'My error message')
    }

    @Test
    public void testWithMandoryParameterDefaultCustomFailureMessageNotProvidedSucceeds() {
        new ConfigurationHelper([myKey: 'myValue']).withMandatoryProperty('myKey')
    }

    @Test
    public void testWithMandoryWithFalseCondition() {
        new ConfigurationHelper([verify: false])
            .withMandatoryProperty('missingKey', null, { c -> return c.get('verify') })
    }

    @Test
    public void testWithMandoryWithTrueConditionMissingValue() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR missingKey')

        new ConfigurationHelper([verify: true])
            .withMandatoryProperty('missingKey', null, { c -> return c.get('verify') })
    }

    @Test
    public void testWithMandoryWithTrueConditionExistingValue() {
        new ConfigurationHelper([existingKey: 'anyValue', verify: true])
            .withMandatoryProperty('existingKey', null, { c -> return c.get('verify') })
    }

    @Test
    public void testTelemetryConfigurationAvailable() {
        Set filter = ['test']
        def configuration = new ConfigurationHelper([test: 'testValue'])
            .mixin([collectTelemetryData: false], filter)
            .use()

        Assert.assertThat(configuration, hasEntry('collectTelemetryData', false))
    }

    @Test
    public void testGStringsAreReplacedByJavaLangStrings() {
        //
        // needed in order to ensure we have real GStrings.
        // a GString not containing variables might be optimized to
        // a java.lang.String from the very beginning.
        def dummy = 'Dummy',
            aGString = "a${dummy}",
            bGString = "b${dummy}",
            cGString = "c${dummy}"

        assert aGString instanceof GString
        assert bGString instanceof GString
        assert cGString instanceof GString

        def config = new ConfigurationHelper([a: aGString,
                                              nextLevel: [b: bGString]])
                     .mixin([c : cGString])
                     .use()

        assert config == [a: 'aDummy',
                          c: 'cDummy',
                       nextLevel: [b: 'bDummy']]
        assert config.a instanceof java.lang.String
        assert config.c instanceof java.lang.String
        assert config.nextLevel.b instanceof java.lang.String
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresAllParamtersArePresentResultsInNoExceptionThrown() {
        new ConfigurationHelper([myKey1: 'a', myKey2: 'b'])
                                   .collectValidationFailures()
                                   .withMandatoryProperty('myKey1')
                                   .withMandatoryProperty('myKey2')
                                   .use()
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresMultipleMissingParametersDoNotResultInFailuresDuringWithMandatoryProperties() {
        new ConfigurationHelper([:]).collectValidationFailures()
                                    .withMandatoryProperty('myKey1')
                                    .withMandatoryProperty('myKey2')
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresMultipleMissingParametersResultsInFailureDuringUse() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR: myKey2, myKey3')
        new ConfigurationHelper([myKey1:'a']).collectValidationFailures()
                                   .withMandatoryProperty('myKey1')
                                   .withMandatoryProperty('myKey2')
                                   .withMandatoryProperty('myKey3')
                                   .use()
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresOneMissingParametersResultsInFailureDuringUse() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR myKey2')
        new ConfigurationHelper([myKey1:'a']).collectValidationFailures()
                                   .withMandatoryProperty('myKey1')
                                   .withMandatoryProperty('myKey2')
                                   .use()
    }

}
