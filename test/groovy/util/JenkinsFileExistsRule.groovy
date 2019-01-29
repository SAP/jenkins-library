package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsFileExistsRule implements TestRule {

    final BasePipelineTest testInstance
    final List existingFiles

    JenkinsFileExistsRule(BasePipelineTest testInstance, List existingFiles) {
        this.testInstance = testInstance
        this.existingFiles = existingFiles
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod('fileExists', [String.class], {s -> return s in existingFiles})

                base.evaluate()
            }
        }
    }

}
