package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.yaml.snakeyaml.Yaml

import static org.junit.Assert.*
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsWriteYamlRule implements TestRule {

    final BasePipelineTest testInstance
    static final String YAML_DUMP = "YAML_DUMP" // key in files map to retrieve the string of an actual Yaml() dump.

    Map files = [:]

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

                    Yaml yaml = new Yaml()
                    StringWriter writer = new StringWriter()
                    yaml.dump(parameterMap.data, writer)

                    // Enable this to actually produce a file.
                    // yaml.dump(parameterMap.data, new FileWriter(parameterMap.file))
                    // yaml.dump(parameterMap.data, new FileWriter("test/resources/variableSubstitution/manifest_out.yml"))

                    files[parameterMap.file] = parameterMap.data
                    files[parameterMap.charset] = parameterMap.charset
                    files[YAML_DUMP] = writer.toString()
                })

                base.evaluate()
            }
        }
    }
}
