package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsEnvironmentRule implements TestRule {
    final BasePipelineTest testInstance

    def env

    JenkinsEnvironmentRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                env = testInstance.loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
                base.evaluate()
            }
        }
    }
}
