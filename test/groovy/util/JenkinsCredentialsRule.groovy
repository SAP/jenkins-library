package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement
import org.jenkinsci.plugins.credentialsbinding.impl.CredentialNotFoundException

/**
 * By default a user &quot;anonymous&quot; with password &quot;********&quot;
 * is provided.
 *
 */
class JenkinsCredentialsRule implements TestRule {

    Map credentials = [:]

    final BasePipelineTest testInstance

    JenkinsCredentialsRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    JenkinsCredentialsRule withCredentials(String credentialsId, String user, String passwd) {
        credentials.put(credentialsId, [user: user, passwd: passwd])
        return this
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {

        return new Statement() {

            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod('usernamePassword', [Map.class],
                    { m -> if (credentials.keySet().contains(m.credentialsId)) return m;
                           // this is what really happens in case of an unknown credentials id,
                           // checked with reality using credentials plugin 2.1.18.
                           throw new CredentialNotFoundException(
                               "Could not find credentials entry with ID '${m.credentialsId}'")
                    })

                testInstance.helper.registerAllowedMethod('withCredentials', [List, Closure], { config, closure ->

                    def credsId = config[0].credentialsId
                    def passwordVariable = config[0].passwordVariable
                    def usernameVariable = config[0].usernameVariable
                    def creds = credentials.get(credsId)

                    binding.setProperty(usernameVariable, creds?.user)
                    binding.setProperty(passwordVariable, creds?.passwd)
                    try {
                        closure()
                    } finally {
                        binding.setProperty(usernameVariable, null)
                        binding.setProperty(passwordVariable, null)
                    }
                })

             base.evaluate()
            }
        }
    }
}
