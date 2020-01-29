package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsReadFileRule implements TestRule {

    final BasePipelineTest testInstance
    final String testRoot
    final Map files = [:]

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

                testInstance.helper.registerAllowedMethod( 'readFile', [String.class], {s -> return load(s, 'UTF-8')} )

                testInstance.helper.registerAllowedMethod( 'readFile', [Map.class], {m -> return load(m.file, m.encoding?m.encoding:'UTF-8')} )

                base.evaluate()
            }
        }
    }

    String load(String path, String encoding) {

        if(files[path]) {
            return files[path]
        }

        if(testRoot == null) {
            throw new IllegalStateException("Test root not set. Resolving: \"${path}\".")
        }
        loadFile(testRoot + '/' + path).getText(encoding)
    }

    File loadFile(String path){
        return new File(path)
    }
}
