package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.yaml.snakeyaml.Yaml

import static org.junit.Assert.*
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsWriteYamlRule implements TestRule {

    final BasePipelineTest testInstance
    static final String DATA = "DATA" // key in files map to retrieve Yaml object graph data..
    static final String CHARSET = "CHARSET" // key in files map to retrieve the charset of the serialized Yaml.
    static final String SERIALIZED_YAML = "SERIALIZED_YAML" // key in files map to retrieve serialized Yaml.

    Map<String, Map<String, Object>> files = new HashMap<>()

    JenkinsWriteYamlRule(BasePipelineTest testInstance) {
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

                testInstance.helper.registerAllowedMethod( 'writeYaml', [Map], { parameterMap ->
                    assertNotNull(parameterMap.file)
                    assertNotNull(parameterMap.data)
                    // charset is optional.

                    Yaml yaml = new Yaml()
                    StringWriter writer = new StringWriter()
                    yaml.dump(parameterMap.data, writer)

                    // Enable this to actually produce a file.
                    // yaml.dump(parameterMap.data, new FileWriter(parameterMap.file))
                    // yaml.dump(parameterMap.data, new FileWriter("test/resources/variableSubstitution/manifest_out.yml"))

                    Map<String, Object> details = new HashMap<>()
                    details.put(DATA, parameterMap.data)
                    details.put(CHARSET, parameterMap.charset ?: "UTF-8")
                    details.put(SERIALIZED_YAML, writer.toString())

                    files[parameterMap.file] = details
                })

                base.evaluate()
            }
        }
    }
}
