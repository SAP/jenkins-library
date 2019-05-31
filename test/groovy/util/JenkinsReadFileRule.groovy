package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsReadFileRule implements TestRule {

    final BasePipelineTest testInstance
    
    // key: file name
    // value: content
    final Map mappings = [:]

    JenkinsReadFileRule(BasePipelineTest testInstance, Map mappings = [:]) {
        this.testInstance = testInstance
        this.mappings << mappings
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod( 'readFile', [String.class], {
                    s ->
                    def content = mappings[s]
                    if(content) return content
                    throw new FileNotFoundException(s)
                })

                testInstance.helper.registerAllowedMethod( 'readFile', [Map.class], {
                    m ->
                    def content = mappings[m.file]
                    if(content) return content
                    throw new FileNotFoundException(m.file)
                })

                base.evaluate()
            }
        }
    }

    public add(String file, String content) {
        mappings.put(file, content)
    }

    public remove(String file) {
        mappings.remove(file)
    }

    File loadFile(String path){
        return new File(path)
    }

}
