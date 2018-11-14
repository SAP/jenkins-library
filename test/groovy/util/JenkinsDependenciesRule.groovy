package util

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.JenkinsShellCallRule.Type

import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

import com.sap.piper.JenkinsUtils

class JenkinsDependenciesRule implements TestRule {

    JenkinsDependenciesRule(BasePipelineTest testInstance) {
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                def preserved = JenkinsUtils.metaClass.static.isPluginActive

                JenkinsUtils.metaClass.static.isPluginActive = { true }

                try {
                    base.evaluate()
                } finally {
                    JenkinsUtils.metaClass.static.isPluginActive = preserved
                }
            }
        }
    }
}
