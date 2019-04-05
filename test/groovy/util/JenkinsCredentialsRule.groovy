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
    Map bindingTypes = [:]

    final BasePipelineTest testInstance

    JenkinsCredentialsRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    JenkinsCredentialsRule withCredentials(String credentialsId, String user, String passwd) {
        credentials.put(credentialsId, [user: user, passwd: passwd])
        return this
    }

    JenkinsCredentialsRule withCredentials(String credentialsId, String token) {
        credentials.put(credentialsId, [token: token])
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
                    { m ->
                                if (credentials.keySet().contains(m.credentialsId)) { bindingTypes[m.credentialsId] = 'usernamePassword'; return m }
                               // this is what really happens in case of an unknown credentials id,
                               // checked with reality using credentials plugin 2.1.18.
                               throw new CredentialNotFoundException(
                                   "Could not find credentials entry with ID '${m.credentialsId}'")
                    })

                testInstance.helper.registerAllowedMethod('string', [Map.class],
                    { m ->
                                if (credentials.keySet().contains(m.credentialsId)) { bindingTypes[m.credentialsId] = 'string'; return m }
                                // this is what really happens in case of an unknown credentials id,
                                // checked with reality using credentials plugin 2.1.18.
                                throw new CredentialNotFoundException(
                                    "Could not find credentials entry with ID '${m.credentialsId}'")
                    })

                testInstance.helper.registerAllowedMethod('withCredentials', [List, Closure], { config, closure ->

                    def credsId = config[0].credentialsId
                    def credentialsBindingType = bindingTypes.get(credsId)
                    def creds = credentials.get(credsId)

                    def tokenVariable, usernameVariable, passwordVariable, prepare, destruct
                    if(credentialsBindingType == "usernamePassword") {
                        passwordVariable = config[0].passwordVariable
                        usernameVariable = config[0].usernameVariable
                        prepare = {
                            binding.setProperty(usernameVariable, creds?.user)
                            binding.setProperty(passwordVariable, creds?.passwd)
                        }
                        destruct = {
                            binding.setProperty(usernameVariable, null)
                            binding.setProperty(passwordVariable, null)
                        }
                    } else if(credentialsBindingType == "string") {
                        tokenVariable = config[0].variable
                        prepare = {
                            binding.setProperty(tokenVariable, creds?.token)
                        }
                        destruct = {
                            binding.setProperty(tokenVariable, null)
                        }
                    } else {
                        throw new RuntimeException("Unknown binding type")
                    }

                    prepare()
                    try {
                        closure()
                    } finally {
                        destruct()
                    }
                })

             base.evaluate()
            }
        }
    }
}
