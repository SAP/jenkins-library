package com.sap.piper

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.yaml.snakeyaml.Yaml

class ConfigurationHelperTest {

    Script mockScript = new Script() {

        def run() {
            // it never runs
            throw new UnsupportedOperationException()
        }

        def STEP_NAME = 'mock'

        def echo(message) {
        }

        def libraryResource(String r) {
            'key: value' // just a stupid default
        }

        def readYaml(Map m) {
           new Yaml().load(m.text)
        }
    }

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

    @Test
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
        Map config = ConfigurationHelper.newInstance(mockScript, [property1: '27']).use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '27'))
    }

    @Test
    void testConfigurationLoaderWithCustomSettings() {
        Map config = ConfigurationHelper.newInstance(mockScript, [property1: '27'])
            .mixin([property1: '41'])
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '41'))
    }

    @Test
    void testConfigurationLoaderWithFilteredCustomSettings() {
        Set filter = ['property2']
        Map config = ConfigurationHelper.newInstance(mockScript, [property1: '27'])
            .mixin([property1: '41', property2: '28', property3: '29'], filter)
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '27'))
        Assert.assertThat(config, hasEntry('property2', '28'))
        Assert.assertThat(config, not(hasKey('property3')))
    }

    @Test
    void testConfigurationHelperLoadingStepDefaults() {
        Set filter = ['property2']
        Map config = ConfigurationHelper.newInstance(mockScript, [property1: '27'])
            .loadStepDefaults()
            .mixinGeneralConfig([configuration:[general: ['general': 'test', 'oldGeneral': 'test2']]], null, [general2: 'oldGeneral'])
            .mixinStageConfig([configuration:[stages:[testStage:['stage': 'test', 'oldStage': 'test2']]]], 'testStage', null, [stage2: 'oldStage'])
            .mixinStepConfig([configuration:[steps:[mock: [step: 'test', 'oldStep': 'test2']]]], null, [step2: 'oldStep'])
            .mixin([property1: '41', property2: '28', property3: '29'], filter)
            .use()
        // asserts
        Assert.assertThat(config, not(hasEntry('property1', '27')))
        Assert.assertThat(config, hasEntry('property2', '28'))
        Assert.assertThat(config, hasEntry('general', 'test'))
        Assert.assertThat(config, hasEntry('general2', 'test2'))
        Assert.assertThat(config, hasEntry('stage', 'test'))
        Assert.assertThat(config, hasEntry('stage2', 'test2'))
        Assert.assertThat(config, hasEntry('step', 'test'))
        Assert.assertThat(config, hasEntry('step2', 'test2'))
        Assert.assertThat(config, not(hasKey('property3')))
    }

    @Test
    void testConfigurationHelperAddIfEmpty() {
        Map config = ConfigurationHelper.newInstance(mockScript, [:])
            .mixin([property1: '41', property2: '28', property3: '29', property4: ''])
            .addIfEmpty('property3', '30')
            .addIfEmpty('property4', '40')
            .addIfEmpty('property5', '50')
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '41'))
        Assert.assertThat(config, hasEntry('property2', '28'))
        Assert.assertThat(config, hasEntry('property3', '29'))
        Assert.assertThat(config, hasEntry('property4', '40'))
    }

    @Test
    void testConfigurationHelperAddIfNull() {
        Map config = ConfigurationHelper.newInstance(mockScript, [:])
            .mixin([property1: '29', property2: '', property3: null])
            .addIfNull('property1', '30')
            .addIfNull('property2', '30')
            .addIfNull('property3', '30')
            .addIfNull('property4', '30')
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', '29'))
        Assert.assertThat(config, hasEntry('property2', ''))
        Assert.assertThat(config, hasEntry('property3', '30'))
        Assert.assertThat(config, hasEntry('property4', '30'))
    }

    @Test
    void testConfigurationHelperDependingOn() {
        Map config = ConfigurationHelper.newInstance(mockScript, [:])
            .mixin([deep: [deeper: 'test'], scanType: 'maven', maven: [path: 'test2']])
            .dependingOn('scanType').mixin('deep/path')
            .use()
        // asserts
        Assert.assertThat(config, hasKey('deep'))
        Assert.assertThat(config.deep, allOf(hasEntry('deeper', 'test'), hasEntry('path', 'test2')))
        Assert.assertThat(config, hasEntry('scanType', 'maven'))
        Assert.assertThat(config, hasKey('maven'))
        Assert.assertThat(config.maven, hasEntry('path', 'test2'))
    }

    @Test
    void testConfigurationHelperWithPropertyInValues() {
        ConfigurationHelper.newInstance(mockScript, [:])
            .mixin([test: 'allowed'])
            .withPropertyInValues('test', ['allowed', 'allowed2'] as Set)
            .use()
    }

    @Test
    void testConfigurationHelperWithPropertyInValuesException() {
        def errorCaught = false
        try {
        ConfigurationHelper.newInstance(mockScript, [:])
            .mixin([test: 'disallowed'])
            .withPropertyInValues('test', ['allowed', 'allowed2'] as Set)
            .use()
        } catch (e) {
            errorCaught = true
            assertThat(e, isA(IllegalArgumentException))
            assertThat(e.getMessage(), is('Invalid test = \'disallowed\'. Valid \'test\' values are: [allowed, allowed2].'))
        }
        assertThat(errorCaught, is(true))
    }

    @Test
    void testConfigurationLoaderWithBooleanValue() {
        Map config = ConfigurationHelper.newInstance(mockScript, [property1: '27'])
            .mixin([property1: false])
            .mixin([property2: false])
            .use()
        // asserts
        Assert.assertThat(config, hasEntry('property1', false))
        Assert.assertThat(config, hasEntry('property2', false))
    }

    @Test
    void testConfigurationLoaderWithMixinDependent() {
        Map config = ConfigurationHelper.newInstance(mockScript, [
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
        def configuration = ConfigurationHelper.newInstance(mockScript)
            .mixin([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'], null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(4))
        Assert.assertThat(configuration.newStructure.new1, is('oldValue1'))
        Assert.assertThat(configuration.newStructure.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityFlat() {
        def configuration = ConfigurationHelper.newInstance(mockScript)
            .mixin([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'], null, [new1: 'old1', new2: 'old2'])
            .use()

        Assert.assertThat(configuration.size(), is(5))
        Assert.assertThat(configuration.new1, is('oldValue1'))
        Assert.assertThat(configuration.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityDeep() {
        def configuration = ConfigurationHelper.newInstance(mockScript)
            .mixin([old1: 'oldValue1', old2: 'oldValue2', test: 'testValue'], null, [deep:[deeper:[newStructure: [new1: 'old1', new2: 'old2']]]])
            .use()

        Assert.assertThat(configuration.size(), is(4))
        Assert.assertThat(configuration.deep.deeper.newStructure.new1, is('oldValue1'))
        Assert.assertThat(configuration.deep.deeper.newStructure.new2, is('oldValue2'))
    }

    @Test
    void testHandleCompatibilityNewAvailable() {
        def configuration = ConfigurationHelper.newInstance(mockScript, [old1: 'oldValue1', newStructure: [new1: 'newValue1'], test: 'testValue'])
            .mixin([old1: 'oldValue1', newStructure: [new1: 'newValue1'], test: 'testValue'], null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(3))
        Assert.assertThat(configuration.newStructure.new1, is('newValue1'))
    }

    @Test
    void testHandleCompatibilityOldNotSet() {
        def configuration = ConfigurationHelper.newInstance(mockScript, [old1: null, test: 'testValue'])
            .mixin([old1: null, test: 'testValue'], null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(2))
        Assert.assertThat(configuration.newStructure.new1, is(null))
    }

    @Test
    void testHandleCompatibilityNoneAvailable() {
        def configuration = ConfigurationHelper.newInstance(mockScript, [old1: null, test: 'testValue'])
            .mixin([test: 'testValue'], null, [newStructure: [new1: 'old1', new2: 'old2']])
            .use()

        Assert.assertThat(configuration.size(), is(2))
        Assert.assertThat(configuration.newStructure.new1, is(null))
    }

    @Test
    void testHandleCompatibilityPremigratedValues() {
        def configuration = ConfigurationHelper.newInstance(mockScript, [old1: null, test: 'testValue'])
            .mixin([someValueToMigrate: 'testValue2'], null, [someValueToMigrateSecondTime: 'someValueToMigrate', newStructure: [new1: 'old1', new2: 'someValueToMigrateSecondTime']])
            .use()

        Assert.assertThat(configuration.size(), is(4))
        Assert.assertThat(configuration.newStructure.new1, is(null))
        Assert.assertThat(configuration.newStructure.new2, is('testValue2'))
    }

    @Test
    public void testWithMandoryParameterReturnDefaultFailureMessage() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR myKey')

        ConfigurationHelper.newInstance(mockScript).withMandatoryProperty('myKey')
    }

    @Test
    public void testWithMandoryParameterReturnCustomerFailureMessage() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('My error message')

        ConfigurationHelper.newInstance(mockScript).withMandatoryProperty('myKey', 'My error message')
    }

    @Test
    public void testWithMandoryParameterDefaultCustomFailureMessageProvidedSucceeds() {
        ConfigurationHelper.newInstance(mockScript, [myKey: 'myValue']).withMandatoryProperty('myKey', 'My error message')
    }

    @Test
    public void testWithMandoryParameterDefaultCustomFailureMessageNotProvidedSucceeds() {
        ConfigurationHelper.newInstance(mockScript, [myKey: 'myValue']).withMandatoryProperty('myKey')
    }

    @Test
    public void testWithMandoryWithFalseCondition() {
        ConfigurationHelper.newInstance(mockScript, [verify: false])
            .withMandatoryProperty('missingKey', null, { c -> return c.get('verify') })
    }

    @Test
    public void testWithMandoryWithTrueConditionMissingValue() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR missingKey')

        ConfigurationHelper.newInstance(mockScript, [verify: true])
            .withMandatoryProperty('missingKey', null, { c -> return c.get('verify') })
    }

    @Test
    public void testWithMandoryWithTrueConditionExistingValue() {
        ConfigurationHelper.newInstance(mockScript, [existingKey: 'anyValue', verify: true])
            .withMandatoryProperty('existingKey', null, { c -> return c.get('verify') })
    }

    @Test
    public void testTelemetryConfigurationAvailable() {
        Set filter = ['test']
        def configuration = ConfigurationHelper.newInstance(mockScript, [test: 'testValue'])
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

        def config = ConfigurationHelper.newInstance(mockScript, [a: aGString,
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
        ConfigurationHelper.newInstance(mockScript, [myKey1: 'a', myKey2: 'b'])
                                   .collectValidationFailures()
                                   .withMandatoryProperty('myKey1')
                                   .withMandatoryProperty('myKey2')
                                   .use()
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresMultipleMissingParametersDoNotResultInFailuresDuringWithMandatoryProperties() {
        ConfigurationHelper.newInstance(mockScript, [:]).collectValidationFailures()
                                    .withMandatoryProperty('myKey1')
                                    .withMandatoryProperty('myKey2')
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresMultipleMissingParametersResultsInFailureDuringUse() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR: myKey2, myKey3')
        ConfigurationHelper.newInstance(mockScript, [myKey1:'a']).collectValidationFailures()
                                   .withMandatoryProperty('myKey1')
                                   .withMandatoryProperty('myKey2')
                                   .withMandatoryProperty('myKey3')
                                   .use()
    }

    @Test
    public void testWithMandatoryParameterCollectFailuresOneMissingParametersResultsInFailureDuringUse() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR myKey2')
        ConfigurationHelper.newInstance(mockScript, [myKey1:'a']).collectValidationFailures()
                                   .withMandatoryProperty('myKey1')
                                   .withMandatoryProperty('myKey2')
                                   .use()
    }

    @Test
    public void testWithPropertyInValuesString() {
        Map config = ['key1':'value1']
        Set possibleValues = ['value1', 'value2', 'value3']

        ConfigurationHelper.newInstance(mockScript, config).collectValidationFailures()
                                   .withPropertyInValues('key1', possibleValues)
                                   .use()
    }

    @Test
    public void testWithPropertyInValuesGString() {
        String value = 'value1'
        Map config = ['key1':"$value"]
        Set possibleValues = ['value1', 'value2', 'value3']

        ConfigurationHelper.newInstance(mockScript, config).collectValidationFailures()
                                   .withPropertyInValues('key1', possibleValues)
                                   .use()
    }

    @Test
    public void testWithPropertyInValuesInt() {
        Map config = ['key1':3]
        Set possibleValues = [1, 2, 3]

        ConfigurationHelper.newInstance(mockScript, config).collectValidationFailures()
                                   .withPropertyInValues('key1', possibleValues)
                                   .use()
    }
}
