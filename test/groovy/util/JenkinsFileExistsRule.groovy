package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import groovy.json.internal.CharacterSource

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsFileExistsRule implements TestRule {

    final BasePipelineTest testInstance
    final Set existingFiles = ['.pipeline/config.yml']

    JenkinsFileExistsRule(BasePipelineTest testInstance, List existingFiles = []) {
        this.testInstance = testInstance
        this.existingFiles.addAll(existingFiles ?: [])
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod('fileExists', [String],
                    {s ->
                        return s in existingFiles
                    })

                base.evaluate()
            }
        }
    }

}
