package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.DefaultValueCache
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement
import org.yaml.snakeyaml.Yaml

class JenkinsReadYamlRule implements TestRule {

    final BasePipelineTest testInstance


    JenkinsReadYamlRule(BasePipelineTest testInstance) {
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
                testInstance.helper.registerAllowedMethod("readYaml", [Map], { Map m ->
                    if(m.text) {
                        return new Yaml().load(m.text)
                    } else if(m.file) {
                        throw new UnsupportedOperationException()
                    } else {
                        throw new IllegalArgumentException("Key 'text' is missing in map ${m}.")
                    }
                })

                base.evaluate()
            }
        }
    }
}
