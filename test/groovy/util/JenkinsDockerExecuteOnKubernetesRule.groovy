package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsDockerExecuteOnKubernetesRule implements TestRule {

    final BasePipelineTest testInstance

    def dockerParams = [:]

    JenkinsDockerExecuteOnKubernetesRule(BasePipelineTest testInstance) {
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

                testInstance.helper.registerAllowedMethod("dockerExecuteOnKubernetes", [Map.class, Closure.class], {map, closure ->
                    dockerParams = map
                    return closure()
                })

                base.evaluate()
            }
        }
    }
}
