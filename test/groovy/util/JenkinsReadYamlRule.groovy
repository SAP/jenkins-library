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


                    return readYaml(yml)
                })

                base.evaluate()
            }
        }
    }

    /**
     * Mimicking code of the original library (link below).
     * <p>
     * Yaml files may contain several YAML sections, separated by ---.
     * This loads them all and returns a {@code List} of entries in case multiple sections were found or just
     * a single {@code Object}, if only one section was read.
     * @see https://github.com/jenkinsci/pipeline-utility-steps-plugin/blob/master/src/main/java/org/jenkinsci/plugins/pipeline/utility/steps/conf/ReadYamlStep.java
     */
    private def readYaml(def yml) {
        Iterable<Object> yaml = new Yaml().loadAll(yml)

        List<Object> result = new LinkedList<Object>()
        for (Object data : yaml) {
            result.add(data)
        }

        // If only one YAML document, return it directly
        if (result.size() == 1) {
            return result.get(0);
        }

        return result;
    }
}
