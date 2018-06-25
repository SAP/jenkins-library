import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.mta.MtaMultiplexer

import groovy.transform.Field

@Field def STEP_NAME = 'snykExecute'

@Field Set GENERAL_CONFIG_KEYS = ['snykCredentialsId']
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'buildDescriptorFile',
    'dockerImage',
    'dockerWorkspace'
    'excludeMtaModules',
    'monitor',
    'scanType',
    'snykOrg',
    'snykResultFile',
    'toJson'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handleStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def utils = parameters.juStabUtils ?: new Utils()
        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            // check mandatory paramerers
            .withMandatoryParameter('dockerImage')
            .withMandatoryParameter('dockerWorkspace')
            .withMandatoryParameter('snykCredentialsId')
            .use()

        utils.unstashAll(config.stashContent)

        switch(config.scanType) {
            case 'mta':
                def scanJobs = [failFast: false]
                // create job for each package.json with scanType: 'npm'
                scanJobs.putAll(MtaMultiplexer.createJobs(
                    this, parameters, config.excludeMtaModules, 'Snyk', 'package.json', 'npm'
                ){options -> executeSnykScan(options)})
                // execute scan jobs in parallel
                parallel scanJobs
                break
            case 'npm':
                // set default file for scanType
                def path = config.buildDescriptorFile.replace('package.json', '')
                try{
                    withCredentials([string(
                        credentialsId: config.snykCredentialsId,
                        variable: 'token'
                    )]) {
                        dockerExecute(
                            dockerImage: config.dockerImage,
                            dockerWorkspace: config.dockerWorkspace,
                            stashContent: config.stashContent,
                            dockerEnvVars: ['SNYK_TOKEN': token]
                        ) {
                            // install Snyk
                            sh "npm install snyk --global --quiet"
                            // install NPM dependencies
                            sh "cd '${path}' && npm install --quiet"
                            // execute Snyk scan
                            def cmd = []
                            cmd.push("cd '${path}'")
                            if(config.monitor) {
                                cmd.push('&& snyk monitor')
                                if(config.snykOrg)
                                    cmd.push("--org=${config.snykOrg}")
                            }
                            cmd.push('&& snyk test')
                            if(config.toJson)
                                cmd.push("--json > ${config.snykResultFile}")
                            sh cmd.join(' ')
                        }
                    }
                }finally{
                    if(config.toJson)
                        archiveArtifacts "${path.replaceAll('\\./', '')}${config.snykResultFile}"
                }
                break
            default:
                error "[ERROR][${STEP_NAME}] ScanType '${config.scanType}' not supported!"
        }
    }
}
