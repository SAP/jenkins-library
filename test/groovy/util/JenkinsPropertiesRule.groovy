package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsPropertiesRule implements TestRule {

    final BasePipelineTest testInstance

    final String propertyPath

    final Properties configProperties

    JenkinsPropertiesRule(BasePipelineTest testInstance, String propertyPath) {
        this.testInstance = testInstance
        this.propertyPath = propertyPath
        configProperties = loadProperties(propertyPath)
    }

    JenkinsPropertiesRule(BasePipelineTest testInstance, String propertyPath, Properties properties) {
        this.testInstance = testInstance
        this.propertyPath = propertyPath
        configProperties = properties
    }


    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("readProperties", [Map.class], {
                    propertyPath ->
                        if (JenkinsPropertiesRule.this.propertyPath.contains(propertyPath.file)) {
                            return JenkinsPropertiesRule.this.configProperties
                        }

                        throw new Exception("Could not find the properties with path $propertyPath")
                })

                base.evaluate()
            }
        }
    }

    static Properties loadProperties(String path) {
        def inputStream = new File(path).newInputStream()
        def properties = new Properties()
        properties.load(inputStream)
        inputStream.close()
        return properties
    }

}
