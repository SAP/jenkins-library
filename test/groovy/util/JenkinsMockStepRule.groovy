package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import java.beans.Introspector

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement


class JenkinsMockStepRule implements TestRule {

    final BasePipelineTest testInstance
    final String stepName
    def callsIndex = 0
    def callsParameters = [:]


    JenkinsMockStepRule(BasePipelineTest testInstance, String stepName) {
        this.testInstance = testInstance
        this.stepName = stepName
    }

    boolean hasParameter(def key, def value){
        for ( def parameters : callsParameters) {
            for ( def parameter : parameters.value.entrySet()) {
                if (parameter.key.equals(key) && parameter.value.equals(value)) return true
            }
        }
        return false
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod(this.stepName, [Map], { Map m ->
                    this.callsIndex += 1
                    this.callsParameters.put(callsIndex, m)
                })

                base.evaluate()
            }
        }
    }

    @Override
    String toString() {
        return callsParameters.toString()
    }

}
