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
            'seleniumExecuteTests',
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
