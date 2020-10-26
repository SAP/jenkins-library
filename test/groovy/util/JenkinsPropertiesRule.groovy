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
                    readPropertyType ->
                        if(readPropertyType.file){
                            if (JenkinsPropertiesRule.this.propertyPath.contains(readPropertyType.file)) {
                            return JenkinsPropertiesRule.this.configProperties
                        }
                    }
                        else if (readPropertyType.text){
                            def propertiesMap = [:]
                            def object = readPropertyType.text.split("=")
                            propertiesMap.put(object[0], object[1])
                            return propertiesMap
                        }

                        throw new Exception("Could not find the properties with path $readPropertyType")
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
