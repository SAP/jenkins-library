import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.mta.MtaMultiplexer
import com.sap.piper.MapUtils

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Credentials for accessing the Snyk API.
     * @possibleValues Jenkins credentials id
     */
    'snykCredentialsId'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * The path to the build descriptor file, e.g. `./package.json`.
     */
    'buildDescriptorFile',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute*/
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute*/
    'dockerWorkspace',
    /**
     * Only scanType 'mta': Exclude modules from MTA projects.
     */
    'exclude',
    /**
     * Monitor the application's dependencies for new vulnerabilities.
     */
    'monitor',
    //TODO: move to general
    /**
     * The type of project that should be scanned.
     * @possibleValues `npm`, `mta`
     */
    'scanType',
    /**
     * Only needed for `monitor: true`: The organisation ID to determine the organisation to report to.
     */
    'snykOrg',
    /**
     * Generate and archive a JSON report.
     */
    'toJson',
    /**
     * Generate and archive a HTML report.
     */
    'toHtml'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

//https://snyk.io/docs/continuous-integration/
/**
 * This step performs an open source vulnerability scan on a *Node project* or *Node module inside an MTA project* through snyk.io.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            // check mandatory parameters
            .withMandatoryProperty('dockerImage')
            .withMandatoryProperty('snykCredentialsId')
            .use()

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
                            script: script,
                            dockerImage: config.dockerImage,
                            dockerEnvVars: MapUtils.merge(['SNYK_TOKEN': token],config.dockerEnvVars?:[:]),
                            dockerWorkspace: config.dockerWorkspace,
                            dockerOptions: config.dockerOptions,
                            stashContent: config.stashContent
                        ) {
                            sh returnStatus: true, script: """
                                node --version
                                npm --version
                            """
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
