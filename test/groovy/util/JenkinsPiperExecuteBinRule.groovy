package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsPiperExecuteBinRule implements TestRule {
    final BasePipelineTest testInstance

    def env

  JenkinsPiperExecuteBinRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                def piperExecuteBin = testInstance.loadScript('piperExecuteBin.groovy').piperExecuteBin
                try {
                    testInstance?.nullScript.piperExecuteBin = piperExecuteBin
                } catch (MissingPropertyException e) {
                    //kept for backward compatibility before all tests inherit from BasePiperTest
                }
                base.evaluate()
            }
        }
    }
}
