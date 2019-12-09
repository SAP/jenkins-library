package util

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.DefaultValueCache

class JenkinsResetDefaultCacheRule implements TestRule {


    JenkinsResetDefaultCacheRule() {
        this(null)
    }

    //
    // Actually not needed. Only provided for the sake of consistency
    // with our other rules which comes with an constructor having the
    // test case contained in the signature.
    JenkinsResetDefaultCacheRule(BasePipelineTest testInstance) {
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                DefaultValueCache.reset()
                base.evaluate()
            }
        }
    }
}
