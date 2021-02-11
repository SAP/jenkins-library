import com.sap.piper.ConfigurationHelper
import com.sap.piper.integration.CloudFoundry
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/newmanExecute.yaml'

@Field Set CONFIG_KEYS = [
    /**
     * Define name array of cloud foundry apps deployed for which secrets (clientid and clientsecret) will be appended
     * to the newman command that overrides the environment json entries
     * (--env-var <appName_clientid>=${clientid} & --env-var <appName_clientsecret>=${clientsecret})
     */
    "cfAppsWithSecrets",
    /**
     * Define an additional repository where the test implementation is located.
     * For protected repositories the `testRepository` needs to contain the ssh git url.
     */
    'testRepository',
    /**
     * Only if `testRepository` is provided: Branch of testRepository, defaults to master.
     */
    'gitBranch',
    /**
     * Only if `testRepository` is provided: Credentials for a protected testRepository
     * @possibleValues Jenkins credentials id
     */
    'gitSshKeyCredentialsId',
]

@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', credentialsId: 'cfCredentialsId', org: 'cfOrg', space: 'cfSpace']]

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixinGeneralConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, CONFIG_KEYS)
        .mixin(parameters, CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
        .use()

    if (parameters.testRepository || config.testRepository ) {
        parameters.stashContent = GitUtils.handleTestRepository(this, [gitBranch: config.gitBranch, gitSshKeyCredentialsId: config.gitSshKeyCredentialsId, testRepository: config.testRepository])
    }

    List<Map> cfCredentials = []
    if (config.cfAppsWithSecrets) {
        CloudFoundry cfUtils = new CloudFoundry(script);
        config.cfAppsWithSecrets.each {
            // def xsuaaCredentials = cfUtils.getXsuaaCredentials(config.cloudFoundry.apiEndpoint,
            //                                                 config.cloudFoundry.org,
            //                                                 config.cloudFoundry.space,
            //                                                 config.cloudFoundry.credentialsId,
            //                                                 appName,
            //                                                 config.verbose ? true : false ) //to avoid config.verbose as "null" if undefined in yaml and since function parameter boolean
            //command_secrets += " --env-var ${appName}_clientid=${xsuaaCredentials.clientid}  --env-var ${appName}_clientsecret=${xsuaaCredentials.clientsecret}"
            def xsuaaCredentials = [clientid: "testClientID", clientsecret: "testClientSecret"]
            cfCredentials.add([var: "PIPER_NEWMANEXECUTE_${appName}_clientid", password: "${xsuaaCredentials.clientid}"])
            cfCredentials.add([var: "PIPER_NEWMANEXECUTE_${appName}_clientsecret", password: "${xsuaaCredentials.clientsecret}"])
            echo "Exposing client id and secret for ${appName}: as ${appName}_clientid and ${appName}_clientsecret to newmanExecute"
            //echo "[INFO]${STEP_NAME}] Preparing credential for being used by piper-go. key: ${it}, exposed as environment variable PIPER_NEWMAN_USER_${it} and PIPER_NEWMAN_PASSWORD_${it}"
            //credentials << [type: 'usernamePassword', id: "${it}", env: ["PIPER_NEWMAN_USER_${it}", "PIPER_NEWMAN_PASSWORD_${it}"], resolveCredentialsId: false]
        }
    }
    print credentials
    withSecretEnv(cfCredentials) {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    }
}

/**
 * Runs code with secret environment variables and hides the values.
 *
 * @param varAndPasswordList - A list of Maps with a 'var' and 'password' key.  Example: `[[var: 'TOKEN', password: 'sekret']]`
 * @param Closure - The code to run in
 * @return {void}
 */
def withSecretEnv(List<Map> varAndPasswordList, Closure closure) {
    wrap([$class: 'MaskPasswordsBuildWrapper', varPasswordPairs: varAndPasswordList]) {
        withEnv(varAndPasswordList.collect { "${it.var}=${it.password}" }) {
            closure()
        }
    }
}
