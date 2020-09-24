package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import org.junit.Assert
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers

class JenkinsLoggingRule implements TestRule {

    final BasePipelineTest testInstance

    def expected = []
    def notExpected = []

    String log = ""

    JenkinsLoggingRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    public JenkinsLoggingRule expect(String substring) {
        expected.add(substring)
        return this
    }

    public JenkinsLoggingRule notExpect(String substring) {
        notExpected.add(substring)
        return this
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("echo", [String.class], {
                    echoInput ->
                        log += "$echoInput \n"
                })

                Throwable caught

                try {
                    base.evaluate()
                } catch(Throwable thr) {
                    caught = thr
                } finally {
                    if(caught instanceof AssertionError) {
                        // Be polite, give other rules the advantage.
                        // We expect other rules located closer to the test case
                        // to throw an AssertionError in case of a violation.
                        throw caught
                    }

                    expected.each { substring -> assertThat("Substring '${substring}' not contained in log.",
                                                            log,
                                                            containsString(substring)) }

                    notExpected.each { substring -> assertThat("Substring '${substring}' is present in log.",
                                                            log,
                                                            not(containsString(substring))) }

                    if(caught != null) {
                        // do not swallow, so that other rules located farer away
                        // to the test case can react
                        throw caught
                    }
                }
            }
        }
    }
}
