import static java.util.stream.Collectors.toList
import static org.hamcrest.Matchers.empty
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat
import static org.junit.Assert.fail

import java.io.File;
import java.util.stream.Collectors
import java.lang.reflect.Field

import org.codehaus.groovy.runtime.metaclass.MethodSelectionException
import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import groovy.io.FileType
import hudson.AbortException
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

/*
 * Intended for collecting generic checks applied to all steps.
 */
public class CommonStepsTest extends BasePiperTest{

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))

    /*
     * With that test we ensure the very first action inside a method body of a call method
     * for a not white listed step is the check for the script handed over properly.
     * Actually we assert for the exception type (AbortException) and for the exception message.
     * In case a new step is added this step will fail. It is the duty of the author of the
     * step to either follow the pattern of checking the script first or to add the step
     * to the white list.
     */
    @Test
    public void scriptReferenceNotHandedOverTest() {

        // all steps not adopting the usual pattern of working with the script.
        def whitelistScriptReference = [
               'commonPipelineEnvironment',
               'handlePipelineStepErrors',
               'pipelineExecute',
               'piperPipeline',
               'prepareDefaultValues',
               'setupCommonPipelineEnvironment',
               'toolValidate',
           ]

        List steps = getSteps().stream()
            .filter {! whitelistScriptReference.contains(it)}
            .forEach {checkReference(it)}
    }

    private void checkReference(step) {

        try {
            def script = loadScript("${step}.groovy")

            try {

                System.setProperty('com.sap.piper.featureFlag.failOnMissingScript', 'true')

                try {
                    script.call([:])
                } catch(AbortException | MissingMethodException e) {
                    throw e
                }  catch(Exception e) {
                    fail "Unexpected exception ${e.getClass().getName()} caught from step '${step}': ${e.getMessage()}"
                }
                fail("Expected AbortException not raised by step '${step}'")

            } catch(MissingMethodException e) {

                // can be improved: exception handling as some kind of control flow.
                // we can also check for the methods and call the appropriate one.

                try {
                    script.call([:]) {}
                } catch(AbortException e1) {
                    throw e1
                }  catch(Exception e1) {
                    fail "Unexpected exception ${e1.getClass().getName()} caught from step '${step}': ${e1.getMessage()}"
                }
                fail("Expected AbortException not raised by step '${step}'")
            }

        } catch(AbortException e) {
            assertThat("Step ''${step} does not fail with expected error message in case mandatory parameter 'script' is not provided.",
                e.getMessage() ==~ /.*\[ERROR\]\[.*\] No reference to surrounding script provided with key 'script', e.g. 'script: this'./,
                is(equalTo(true)))
        } finally {
            System.clearProperty('com.sap.piper.featureFlag.failOnMissingScript')
        }
    }

    private static fieldRelatedWhitelist = [
            'toolValidate', // step is intended to be configured by other steps
            'durationMeasure', // only expects parameters via signature
            'prepareDefaultValues', // special step (infrastructure)
            'piperPipeline', // special step (infrastructure)
            'pipelineStashFilesAfterBuild', // intended to be called from pipelineStashFiles
            'pipelineStashFilesBeforeBuild', // intended to be called from pipelineStashFiles
            'pipelineStashFiles', // only forwards to before/after step
            'pipelineExecute', // special step (infrastructure)
            'commonPipelineEnvironment', // special step (infrastructure)
            'handlePipelineStepErrors', // special step (infrastructure)
            'piperStageWrapper' //intended to be called from within stages
            ]

    @Test
    public void generalConfigKeysSetPresentTest() {

        def fieldName = 'GENERAL_CONFIG_KEYS'
        // the steps added to the fieldRelatedWhitelist do not take the general config at all
        def stepsWithoutGeneralConfigKeySet = fieldCheck(fieldName, fieldRelatedWhitelist.plus(['gaugeExecuteTests',
                                                                                                'pipelineRestartSteps']))

        assertThat("Steps without ${fieldName} field (or that field is not a Set): ${stepsWithoutGeneralConfigKeySet}",
            stepsWithoutGeneralConfigKeySet, is(empty()))
    }

    @Test
    public void stepConfigKeysSetPresentTest() {

        def fieldName = 'STEP_CONFIG_KEYS'
        def stepsWithoutStepConfigKeySet = fieldCheck(fieldName, fieldRelatedWhitelist.plus('setupCommonPipelineEnvironment'))

        assertThat("Steps without ${fieldName} field (or that field is not a Set): ${stepsWithoutStepConfigKeySet}",
            stepsWithoutStepConfigKeySet, is(empty()))
    }

    @Test
    public void parametersKeysSetPresentTest() {

        def fieldName = 'PARAMETER_KEYS'
        def stepsWithoutParametersKeySet = fieldCheck(fieldName, fieldRelatedWhitelist.plus('setupCommonPipelineEnvironment'))

        assertThat("Steps without ${fieldName} field (or that field is not a Set): ${stepsWithoutParametersKeySet}",
            stepsWithoutParametersKeySet, is(empty()))
    }

    private fieldCheck(fieldName, whitelist) {

        def stepsWithoutGeneralConfigKeySet = []

        for(def step in getSteps()) {
            if(whitelist.contains(step)) continue

            def fields = loadScript("${step}.groovy").getClass().getDeclaredFields() as Set
            Field generalConfigKeyField = fields.find{ it.getName() == fieldName}
            if(! generalConfigKeyField ||
               ! generalConfigKeyField
                   .getType()
                   .isAssignableFrom(Set.class)) {
                        stepsWithoutGeneralConfigKeySet.add(step)
            }
        }
        return stepsWithoutGeneralConfigKeySet
    }

    @Test
    public void stepsWithWrongFieldNameTest() {

        def whitelist = [
            'commonPipelineEnvironment',
            'piperPipeline'
        ]

        def stepsWithWrongStepName = []

        for(def step in getSteps()) {

            if(whitelist.contains(step)) continue

            def script = loadScript("${step}.groovy")

            def fields = script.getClass().getDeclaredFields() as Set
            Field stepNameField = fields.find { it.getName() == 'STEP_NAME'}

            if(! stepNameField) {
                stepsWithWrongStepName.add(step)
                continue
            }

            boolean notAccessible = false;
            def fieldName

            if(!stepNameField.isAccessible()) {
                stepNameField.setAccessible(true)
                notAccessible = true
            }

            try {
                fieldName = stepNameField.get(script)
            } finally {
                if(notAccessible) stepNameField.setAccessible(false)
            }
            if(fieldName != step) {
                stepsWithWrongStepName.add(step)
            }
        }

        assertThat("Steps with wrong step name or without STEP_NAME field.: ${stepsWithWrongStepName}",
            stepsWithWrongStepName, is(empty()))
    }

    /*
     * With that test we ensure that all return types of the call methods of all the steps
     * are void. Return types other than void are not possible when running inside declarative
     * pipelines. Parameters shared between several steps needs to be shared via the commonPipelineEnvironment.
     */
    @Test
    public void returnTypeForCallMethodsIsVoidTest() {

        def stepsWithCallMethodsOtherThanVoid = []

        def whitelist = [
            'durationMeasure',
            ]

        for(def step in getSteps()) {
            def methods = loadScript("${step}.groovy").getClass().getDeclaredMethods() as List
            Collection callMethodsWithReturnTypeOtherThanVoid =
                methods.stream()
                       .filter { ! whitelist.contains(step) }
                       .filter { it.getName() == 'call' &&
                                 it.getReturnType() != Void.TYPE }
                       .collect(toList())
            if(!callMethodsWithReturnTypeOtherThanVoid.isEmpty()) stepsWithCallMethodsOtherThanVoid << step
        }

        assertThat("Steps with call methods with return types other than void: ${stepsWithCallMethodsOtherThanVoid}",
            stepsWithCallMethodsOtherThanVoid, is(empty()))
    }

    private static getSteps() {
        List steps = []
        new File('vars').traverse(type: FileType.FILES, maxDepth: 0)
            { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars[\\\/](.*)\.groovy/)[0][1] }
        return steps
    }
}
