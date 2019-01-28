package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class JenkinsLockRule implements TestRule {

    final BasePipelineTest testInstance
    final List lockResources = []

    JenkinsLockRule(BasePipelineTest testInstance) {
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

                testInstance.helper.registerAllowedMethod("lock", [String.class, Closure.class], {
                    resource, body ->
                        lockResources.add(resource)
                        body()
                })

                base.evaluate()
            }
        }
    }
}
