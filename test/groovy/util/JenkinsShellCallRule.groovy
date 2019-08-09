package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsShellCallRule implements TestRule {

    enum Type { PLAIN, REGEX }

    class Command {

        final Type type
        final String script

        Command(Type type, String script) {
            this.type = type
            this.script = script
        }

        String toString() {
            return "${type} : ${script}"
        }

        @Override
        public int hashCode() {
            return type.hashCode() * script.hashCode()
        }

        @Override
        public boolean equals(Object obj) {

            if (obj == null || !obj instanceof Command) return false
            Command other = (Command) obj
            return type == other.type && script == other.script
        }
    }

    final BasePipelineTest testInstance

    List shell = []

    Map<Command, String> returnValues = [:]

    JenkinsShellCallRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    def setReturnValue(script, value) {
        setReturnValue(Type.PLAIN, script, value)
    }

    def setReturnValue(type, script, value) {
        returnValues[new Command(type, script)] = value
    }

    def handleShellCall(Map parameters) {

        def unifiedScript = unify(parameters.script)

        shell.add(unifiedScript)

        def result = null

        for(def e : returnValues.entrySet()) {
            if(e.key.type == Type.REGEX && unifiedScript =~ e.key.script) {
                result =  e.value
                break
        } else if(e.key.type == Type.PLAIN && unifiedScript.equals(e.key.script)) {
                result = e.value
                break
            }
        }
        if(result instanceof Closure) result = result()
        if (!result && parameters.returnStatus) result = 0

        if(! parameters.returnStdout && ! parameters.returnStatus) return
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

                testInstance.helper.registerAllowedMethod("sh", [String.class], {
                    command -> handleShellCall([
                        script: command,
                        returnStdout: false,
                        returnStatus: false
                    ])
                })

                testInstance.helper.registerAllowedMethod("sh", [Map.class], {
                    m -> handleShellCall(m)
                })

                base.evaluate()
            }
        }
    }

    private static String unify(String s) {
        s.replaceAll(/\s+/," ").trim()
    }
}
