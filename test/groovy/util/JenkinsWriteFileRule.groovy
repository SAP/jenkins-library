package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertTrue
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsWriteFileRule implements TestRule {

    final BasePipelineTest testInstance

    Map files = [:]

    JenkinsWriteFileRule(BasePipelineTest testInstance) {
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

                testInstance.helper.registerAllowedMethod( 'writeFile', [Map.class], { m ->
                    assertNotNull(m.file)
                    assertTrue(m.file instanceof CharSequence)
                    assertNotNull(m.text)
                    assertTrue(m.text instanceof CharSequence)
                    if (m.encoding) {
                        assertTrue(m.encoding instanceof CharSequence)
                        // Would be nice to actually handle encoding
                    }
                    files[m.file] = m.text
                })

                base.evaluate()
            }
        }
    }
}
