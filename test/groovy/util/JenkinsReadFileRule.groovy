package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsReadFileRule implements TestRule {

    final BasePipelineTest testInstance
    final String testRoot

    JenkinsReadFileRule(BasePipelineTest testInstance, String testRoot) {
        this.testInstance = testInstance
        this.testRoot = testRoot
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod( 'readFile', [String.class], {s -> return (loadFile("${testRoot}/${s}")).getText('UTF-8')} )

                testInstance.helper.registerAllowedMethod( 'readFile', [Map.class], {m -> return (loadFile("${testRoot}/${m.file}")).getText(m.encoding?m.encoding:'UTF-8')} )

                base.evaluate()
            }
        }
    }

    File loadFile(String path){
        return new File(path)
    }

}
