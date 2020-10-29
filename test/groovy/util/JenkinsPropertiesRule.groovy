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
                            throw new Exception("Could not find the properties with path $readPropertyType")
                        }
                        else if (readPropertyType.text){
                            // Multiline properties are not supported.
                            def propertiesMap = [:]
                            for (def line : new StringReader(readPropertyType.text)) {
                                if (! line.trim()) continue
                                entry = line.split('=')
                                if(entry.length != 2) {
                                    throw new RuntimeException("Invalid properties: ${readPropertyType.text}. Line '${line}' does not contain a valid key value pair ")
                                }
                                propertiesMap.put(entry[0], entry[1])

                            }
                            return propertiesMap
                        }
                        throw new Exception("neither 'text' nor 'file' argument was provided to 'readProperties'")
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
