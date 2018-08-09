package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsWriteJsonRule implements TestRule {

    final BasePipelineTest testInstance

    Map files = [:]

    JenkinsWriteJsonRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod( 'writeJSON', [Map.class], {m -> files[m.file] = m.json.toString()})

                base.evaluate()
            }
        }
    }
}
