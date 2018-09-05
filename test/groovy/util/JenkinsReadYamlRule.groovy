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

    JenkinsReadYamlRule registerYaml(fileName, yaml) {
        ymls.put(fileName, yaml)
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
                    def yml
                    if(m.text) {
                        yml = m.text
                    } else if(m.file) {
                        yml = ymls.get(m.file)
                        if(yml == null) throw new NullPointerException("yaml file '${m.file}' not registered.")
                        if(yml instanceof Closure) yml = yml()
                    } else {
                        throw new IllegalArgumentException("Key 'text' and 'file' are both missing in map ${m}.")
                    }
                    return new Yaml().load(yml)
                })

                base.evaluate()
            }
        }
    }
}
