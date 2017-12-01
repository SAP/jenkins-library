package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsScriptLoaderRule implements TestRule {

    final BasePipelineTest testInstance

    final String scriptBasePath

    JenkinsScriptLoaderRule(BasePipelineTest testInstance, String scriptBasePath) {
        this.testInstance = testInstance
        this.scriptBasePath = scriptBasePath
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("load", [String.class], {
                    fileNameIntegration ->
                        return testInstance.loadScript("$scriptBasePath/$fileNameIntegration")
                })

                base.evaluate()
            }
        }
    }
}
