import com.sap.piper.ConfigurationHelper
import com.sap.piper.integration.CloudFoundry
import groovy.transform.Field
import com.sap.piper.GitUtils

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
        parameters.stashContent = [GitUtils.handleTestRepository(this, [gitBranch: config.gitBranch, gitSshKeyCredentialsId: config.gitSshKeyCredentialsId, testRepository: config.testRepository])]
    }

    List<String> cfCredentials = []
    if (config.cfAppsWithSecrets) {
        CloudFoundry cfUtils = new CloudFoundry(script);
        config.cfAppsWithSecrets.each { appName ->
            def xsuaaCredentials = cfUtils.getXsuaaCredentials(config.cloudFoundry.apiEndpoint,
                                                            config.cloudFoundry.org,
                                                            config.cloudFoundry.space,
                                                            config.cloudFoundry.credentialsId,
                                                            appName,
                                                            config.verbose ? true : false )
            cfCredentials.add("PIPER_NEWMANEXECUTE_${appName}_clientid=${xsuaaCredentials.clientid}")
            cfCredentials.add("PIPER_NEWMANEXECUTE_${appName}_clientsecret=${xsuaaCredentials.clientsecret}")
            echo "Exposing client id and secret for ${appName}: as ${appName}_clientid and ${appName}_clientsecret to newmanExecute"
        }
    }
    withEnv(cfCredentials) {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
    }
}
