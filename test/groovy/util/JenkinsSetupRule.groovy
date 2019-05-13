package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsSetupRule implements TestRule {

    def library = SharedLibraryCreator.implicitLoadedLibrary

    final BasePipelineTest testInstance

    JenkinsSetupRule(BasePipelineTest testInstance) {
        this(testInstance, null)
    }

    JenkinsSetupRule(BasePipelineTest testInstance, LibraryConfiguration configuration) {
        this.testInstance = testInstance
        if(configuration)
            this.library = configuration
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.scriptRoots += "vars/"
                testInstance.setUp()
                // register library
                testInstance.helper.registerSharedLibrary(library)
                // set jenkins job mock variables
                testInstance.binding.setVariable('env', [
                    JOB_NAME    : 'p',
                    BUILD_NUMBER: '1',
                    BUILD_URL   : 'http://build.url',
                    BRANCH_NAME: 'master',
                    WORKSPACE: 'any/path'
                ])

                base.evaluate()

            }
        }
    }
}
