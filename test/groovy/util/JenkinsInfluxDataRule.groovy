package util

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.analytics.InfluxData

class JenkinsInfluxDataRule implements TestRule {
    JenkinsInfluxDataRule() { this(null) }

    // Actually not needed. Only provided for the sake of consistency
    // with our other rules which comes with an constructor having the
    // test case contained in the signature.
    JenkinsInfluxDataRule(BasePipelineTest testInstance) {}

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                InfluxData.reset()
                base.evaluate()
            }
        }
    }
}
