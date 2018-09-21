import static java.util.stream.Collectors.toList
import static org.hamcrest.Matchers.empty
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat
import static org.junit.Assert.fail

import java.io.File;
import java.util.stream.Collectors

import org.codehaus.groovy.runtime.metaclass.MethodSelectionException
import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Rule
import org.junit.Test;
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import groovy.io.FileType
import hudson.AbortException
import util.BasePiperTest
import util.JenkinsStepRule
import util.Rules

/*
 * Intended for collecting generic checks applied to all steps.
 */
public class CommonStepsTest extends BasePiperTest{

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)

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
               'prepareDefaultValues',
               'setupCommonPipelineEnvironment',
               'toolValidate',
           ]

        List steps = getSteps().stream()
            .filter {! whitelistScriptReference.contains(it)}
            .forEach {checkReference(it)}
    }

    private static List getSteps() {
        List steps = []
        new File('vars').traverse(type: FileType.FILES, maxDepth: 0)
            { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars\/(.*)\.groovy/)[0][1] }
        return steps

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
}
