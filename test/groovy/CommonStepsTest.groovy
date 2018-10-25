import static java.util.stream.Collectors.toList
import static org.hamcrest.Matchers.empty
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat
import static org.junit.Assert.fail

import java.lang.reflect.Field
import java.io.File;

import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import groovy.io.FileType
import util.BasePiperTest

import util.Rules

/*
 * Intended for collecting generic checks applied to all steps.
 */
public class CommonStepsTest extends BasePiperTest{

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)

    private static fieldRelatedWhitelist = [
            'toolValidate', // step is intended to be configured by other steps
            'durationMeasure', // only expects parameters via signature
            'prepareDefaultValues', // special step (infrastructure)
            'pipelineStashFilesAfterBuild', // intended to be called from pipelineStashFiles
            'pipelineStashFilesBeforeBuild', // intended to be called from pipelineStashFiles
            'pipelineStashFiles', // only forwards to before/after step
            'pipelineExecute', // special step (infrastructure)
            'prepareDefaultValues', // special step (infrastructure)
            'commonPipelineEnvironment', // special step (infrastructure)
            'handlePipelineStepErrors', // special step (infrastructure)
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

        def whitelist = ['commonPipelineEnvironment']

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
            'transportRequestCreate',
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
            { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars\/(.*)\.groovy/)[0][1] }
        return steps

    }
}
