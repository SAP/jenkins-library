package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import groovy.json.JsonSlurper
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsReadJsonRule implements TestRule {

    final BasePipelineTest testInstance
    final String testRoot

    JenkinsReadJsonRule(BasePipelineTest testInstance, testRoot = '') {
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
                testInstance.helper.registerAllowedMethod("readJSON", [Map], { Map m ->
                    if(m.text) {
                        def js = new JsonSlurper()
                        return js.parseText(m.text)
                    } else if(m.file) {
                        def js = new JsonSlurper()
                        def reader = new BufferedReader(new FileReader( "${this.testRoot}${m.file}" ))
                        return js.parse(reader)
                    } else {
                        throw new IllegalArgumentException("Key 'text' is missing in map ${m}.")
                    }
                })

                base.evaluate()
            }
        }
    }
}
