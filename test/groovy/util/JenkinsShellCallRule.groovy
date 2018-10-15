package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.JenkinsShellCallRule.Type

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsShellCallRule implements TestRule {

    enum Type { PLAIN, REGEX }

    class Key{

        final Type type
        final String script

        Key(Type type, String script) {
            this.type = type
            this.script = script
        }
    }

    final BasePipelineTest testInstance

    List shell = []

    Map<Key, String> returnValues = [:]
    Map<Key, String> returnStatus = [:]


    JenkinsShellCallRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    def setReturnValue(script, value) {
        setReturnValue(Type.PLAIN, script, value)
    }

    def setReturnValue(type, script, value) {
        returnValues[new Key(type, script)] = value
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
                    command ->
                        shell.add(unify(command))
                })

                testInstance.helper.registerAllowedMethod("sh", [Map.class], {
                    m ->
                        shell.add(m.script.replaceAll(/\s+/," ").trim())
                        if (m.returnStdout || m.returnStatus) {
                            def unifiedScript = unify(m.script)
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
                            if (!result && m.returnStatus) result = 0
                            return result
                        }
                })

                base.evaluate()
            }
        }
    }

    private static String unify(String s) {
        s.replaceAll(/\s+/," ").trim()
    }
}
