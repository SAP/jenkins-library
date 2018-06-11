package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsShellCallRule implements TestRule {

    final BasePipelineTest testInstance

    List shell = []

    def returnValues = [:]

    JenkinsShellCallRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    def setReturnValue(script, value) {
        returnValues[script] = value
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("sh", [String.class], {
                    command ->
                        shell.add(unify(command))
                })

                testInstance.helper.registerAllowedMethod("sh", [Map.class], {
                    m ->
                        shell.add(m.script.replaceAll(/\s+/," ").trim())
                        if (m.returnStdout || m.returnStatus)
                            return returnValues[unify(m.script)]
                })

                base.evaluate()
            }
        }
    }

    private static String unify(String s) {
        s.replaceAll(/\s+/," ").trim()
    }
}
