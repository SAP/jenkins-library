package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.DefaultValueCache
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement
import org.yaml.snakeyaml.Yaml

class JenkinsConfigRule implements TestRule {

    final BasePipelineTest testInstance


    JenkinsConfigRule(BasePipelineTest testInstance) {
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
				testInstance.helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
					Yaml yamlParser = new Yaml()
					return yamlParser.load(parameters.text)
				})
				
				DefaultValueCache.reset()

                base.evaluate()
            }
        }
    }
}
