package util

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement
import org.yaml.snakeyaml.Yaml

import com.lesfurets.jenkins.unit.BasePipelineTest

class JenkinsReadYamlRule implements TestRule {
    final BasePipelineTest testInstance

    // Empty project configuration file registered by default
    // since almost every test needs it.
    def ymls = ['.pipeline/config.yml': {''}]

    JenkinsReadYamlRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    JenkinsReadYamlRule registerYaml(fileName, closure) {
        ymls.put(fileName, closure)
        return this
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
                        def closure = ymls.get(m.file)
                        if(!closure) throw new NullPointerException("yaml file '${m.file}' not registered.")
                        return new Yaml().load(closure())
                    } else {
                        throw new IllegalArgumentException("Key 'text' is missing in map ${m}.")
                    }
                })

                base.evaluate()
            }
        }
    }
}
