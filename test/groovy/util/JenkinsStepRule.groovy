package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import java.beans.Introspector

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement


class JenkinsStepRule implements TestRule {

    final BasePipelineTest testInstance

    def step

    JenkinsStepRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                def testClassName = testInstance.getClass().getSimpleName()
                def stepName = Introspector.decapitalize(testClassName.replaceAll('Test$', ''))
                this.step = testInstance.loadScript("${stepName}.groovy")

                base.evaluate()
            }
        }
    }
}
