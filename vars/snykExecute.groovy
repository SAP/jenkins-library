import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.mta.MtaMultiplexer

import groovy.transform.Field

@Field def STEP_NAME = 'snykExecute'

@Field Set GENERAL_CONFIG_KEYS = ['snykCredentialsId']
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'buildDescriptorFile',
    'dockerImage',
    'exclude',
    'monitor',
    'scanType',
    'snykOrg',
    'toJson',
    'toHtml'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters?.juStabUtils ?: new Utils()
        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            // check mandatory paramerers
            .withMandatoryProperty('dockerImage')
            .withMandatoryProperty('snykCredentialsId')
            .use()

        new Utils().pushToSWA([step: STEP_NAME], config)

        utils.unstashAll(config.stashContent)

        switch(config.scanType) {
            case 'mta':
                def scanJobs = [failFast: false]
                // create job for each package.json with scanType: 'npm'
                scanJobs.putAll(MtaMultiplexer.createJobs(
                    this, parameters, config.exclude, 'Snyk', 'package.json', 'npm'
                ){options -> snykExecute(options)})
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
                            stashContent: config.stashContent,
                            dockerEnvVars: ['SNYK_TOKEN': token]
                        ) {
                            // install Snyk
                            sh 'npm install snyk --global --quiet'
                            if(config.toHtml){
                                config.toJson = true
                                sh 'npm install snyk-to-html --global --quiet'
                            }
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
                                cmd.push("--json > snyk.json")
                            try{
                                sh cmd.join(' ')
                            }finally{
                                if(config.toHtml) sh "snyk-to-html -i ${path}snyk.json -o ${path}snyk.html"
                            }
                        }
                    }
                }finally{
                    if(config.toJson) archiveArtifacts "${path.replaceAll('\\./', '')}snyk.json"
                    if(config.toHtml) archiveArtifacts "${path.replaceAll('\\./', '')}snyk.html"
                }
                break
            default:
                error "[ERROR][${STEP_NAME}] ScanType '${config.scanType}' not supported!"
        }
    }
}
