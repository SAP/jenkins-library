package util

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

import com.lesfurets.jenkins.unit.BasePipelineTest

class JenkinsLibraryResourceRule implements TestRule {

    final BasePipelineTest testInstance

    JenkinsLibraryResourceRule(BasePipelineTest testInstance) {
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
                testInstance.helper.registerAllowedMethod("libraryResource", [String], { r ->
                    File resource = new File(new File('resources'), r)
                    if(! resource.exists()) {
                        throw new RuntimeException("Resource '${resource}' not found.")
                    }
                    return resource.text
                })

                base.evaluate()
            }
        }
    }
}
