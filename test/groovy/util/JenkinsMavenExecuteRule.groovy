package groovy.util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsMavenExecuteRule implements TestRule {

    final BasePipelineTest testInstance

    List executions = []

    Map<String, String> returnValues = [:]

    JenkinsMavenExecuteRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    def setReturnValue(String params, String value) {
        returnValues.put(params, value)
    }

    def handleExecution(Map parameters) {

        String params = stringify(parameters)
        executions.add(params)

        def result = returnValues.get(params)

        for (def e : returnValues.entrySet()) {
            if (params == e.key.params) {
                result = e.value
                break
            }
        }
        if (result instanceof Closure) {
            result = result()
        }
        if (!result && parameters.returnStatus) {
            result = 0
        }

        if (!parameters.returnStdout && !parameters.returnStatus) {
            return
        }
        return result
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("mavenExecute", [Map.class], {
                    map -> handleExecution(map)
                })

                base.evaluate()
            }
        }
    }

    private static String stringify(Map map) {
        String params = ''
        List keys = ['pomPath', 'goals', 'defines', 'flags']
        for (String key : keys) {
            if (params.length() > 0)
                params += ' '
            params += map.get(key)
        }
        return params.replaceAll(/\s+/," ").trim()
    }
}
