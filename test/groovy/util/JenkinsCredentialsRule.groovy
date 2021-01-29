package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import groovy.json.JsonSlurper
import org.jenkinsci.plugins.credentialsbinding.impl.CredentialNotFoundException
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

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

    JenkinsCredentialsRule withCredentials(String credentialsId, String secretTextOrFilePath) {
            credentials.put(credentialsId, [token: secretTextOrFilePath])
            return this
    }

    JenkinsCredentialsRule reset(){
        credentials.clear()
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

                testInstance.helper.registerAllowedMethod('file', [Map.class],
                    { m ->
                        if (credentials.keySet().contains(m.credentialsId)) { bindingTypes[m.credentialsId] = 'file'; return m }
                        // this is what really happens in case of an unknown credentials id,
                        // checked with reality using credentials plugin 2.1.18.
                        throw new CredentialNotFoundException(
                            "Could not find credentials entry with ID '${m.credentialsId}'")
                    })

                testInstance.helper.registerAllowedMethod('withCredentials', [List, Closure], { config, closure ->
                    // there can be multiple credentials defined for the closure; collecting the necessary binding
                    // preparations and destructions before executing closure
                    def preparations = []
                    def destructions = []
                    config.each { cred ->
                        def credsId = cred.credentialsId
                        def credentialsBindingType = bindingTypes.get(credsId)
                        def creds = credentials.get(credsId)

                        def tokenVariable, usernameVariable, passwordVariable, prepare, destruct
                        if (credentialsBindingType == "usernamePassword") {
                            passwordVariable = cred.passwordVariable
                            usernameVariable = cred.usernameVariable
                            preparations.add({
                                binding.setProperty(usernameVariable, creds?.user)
                                binding.setProperty(passwordVariable, creds?.passwd)
                            })
                            destructions.add({
                                binding.setProperty(usernameVariable, null)
                                binding.setProperty(passwordVariable, null)
                            })
                        } else if (credentialsBindingType == "string") {
                            tokenVariable = cred.variable
                            preparations.add({
                                binding.setProperty(tokenVariable, creds?.token)
                            })
                            destructions.add({
                                binding.setProperty(tokenVariable, null)
                            })
                        }
                        else if (credentialsBindingType == "file") {
                            fileContentVariable = cred.variable
                            preparations.add({
                                binding.setProperty(fileContentVariable, creds?.token)
                            })
                            destructions.add({
                                binding.setProperty(fileContentVariable, null)
                            })
                        }
                        else {
                            throw new RuntimeException("Unknown binding type")
                        }
                    }

                    preparations.each { it() }
                    try {
                        closure()
                    } finally {
                        destructions.each { it() }
                    }
                })

                base.evaluate()
            }
        }
    }
}
