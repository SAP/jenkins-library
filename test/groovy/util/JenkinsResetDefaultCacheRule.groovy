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

                MetaMethod oldReadDefaults = DefaultValueCache.metaClass.
                    getStaticMetaMethod("readDefaults", Script)
                MetaMethod oldPersistDefaults = DefaultValueCache.metaClass.
                    getStaticMetaMethod("persistDefaults", [Script, Map, List])

                DefaultValueCache.metaClass.static.readDefaults = { Script s ->
                    return null
                }
                DefaultValueCache.metaClass.static.persistDefaults = { Script s, Map dv, List cd ->
                    return null
                }

                base.evaluate()

                DefaultValueCache.metaClass.static.readDefaults = oldReadDefaults
                DefaultValueCache.metaClass.static.persistDefaults = oldPersistDefaults
            }
        }
    }
}
