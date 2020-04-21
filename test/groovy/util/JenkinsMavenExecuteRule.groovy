package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsMavenExecuteRule implements TestRule {

    static class Execution {

        final String pomPath
        final List goals
        final List defines
        final List flags

        Execution(Map parameters) {
            this.pomPath = parameters.pomPath ?: 'pom.xml'
            this.goals = asList(parameters.goals)
            this.defines = asList(parameters.defines)
            this.flags = asList(parameters.flags)
        }

        String toString() {
            return "--file ${pomPath} : ${goals} : ${defines} : ${flags}"
        }

        @Override
        int hashCode() {
            return pomPath.hashCode() * goals.hashCode() * defines.hashCode() * flags.hashCode()
        }

        @Override
        boolean equals(Object obj) {
            if (obj == null || !obj instanceof Execution) {
                return false
            }
            Execution other = (Execution) obj
            return goals == other.goals && defines == other.defines && flags == other.flags
        }

        private List asList(def value) {
            if (value instanceof List) {
                return value as List
            }
            if (value instanceof CharSequence) {
                return [ value ]
            }
            return []
        }
    }

    final BasePipelineTest testInstance

    List<Execution> executions = []

    Map<String, String> returnValues = [:]

    JenkinsMavenExecuteRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    def setReturnValue(Map params, String value) {
        returnValues.put(stringify(params), value)
    }

    def handleExecution(Map parameters) {

        String params = stringify(parameters)
        executions.add(new Execution(parameters))

        def result = returnValues.get(params)
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
