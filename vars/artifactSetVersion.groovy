import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'artifactSetVersion'
@Field Set STEP_CONFIG_KEYS = [
    'artifactType',
    'buildTool',
    'commitVersion',
    'dockerVersionSource',
    'filePath',
    'gitCredentialsId',
    'gitUserEMail',
    'gitUserName',
    'gitSshUrl',
    'tagPrefix',
    'timestamp',
    'timestampTemplate',
    'versioningTemplate'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus('gitCommitId')

def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def gitUtils = parameters.juStabGitUtils
        if (gitUtils == null) {
            gitUtils = new GitUtils()
        }

        if (fileExists('.git')) {
            if (sh(returnStatus: true, script: 'git diff --quiet HEAD') != 0)
                error "[${STEP_NAME}] Files in the workspace have been changed previously - aborting ${STEP_NAME}"
        }

        def script = parameters.script
        if (script == null)
            script = this

        // load default & individual configuration
        Map configuration = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(gitCommitId: gitUtils.getGitCommitIdOrNull())
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        def utils = new Utils()
        def buildTool = utils.getMandatoryParameter(configuration, 'buildTool')

        if (!configuration.filePath)
            configuration.filePath = configuration[buildTool].filePath //use default configuration

        def newVersion
        def artifactVersioning = ArtifactVersioning.getArtifactVersioning(buildTool, script, configuration)

        if(configuration.artifactType == 'appContainer' && configuration.dockerVersionSource == 'appVersion'){
            if (script.commonPipelineEnvironment.getArtifactVersion())
                //replace + sign if available since + is not allowed in a Docker tag
                newVersion = script.commonPipelineEnvironment.getArtifactVersion().replace('+', '_')
            else
                error ("[${STEP_NAME}] No artifact version available for 'dockerVersionSource: appVersion' -> executeBuild needs to run for the application artifact first to set the artifactVersion for the application artifact.'")
        } else {
            def currentVersion = artifactVersioning.getVersion()

            def timestamp = configuration.timestamp ? configuration.timestamp : getTimestamp(configuration.timestampTemplate)

            def versioningTemplate = configuration.versioningTemplate ? configuration.versioningTemplate : configuration[configuration.buildTool].versioningTemplate
            //defined in default configuration
            def binding = [version: currentVersion, timestamp: timestamp, commitId: configuration.gitCommitId]
            def templatingEngine = new SimpleTemplateEngine()
            def template = templatingEngine.createTemplate(versioningTemplate).make(binding)
            newVersion = template.toString()
        }

        artifactVersioning.setVersion(newVersion)

        def gitCommitId

        if (configuration.commitVersion) {
            sh 'git add .'

            sshagent([configuration.gitCredentialsId]) {
                def gitUserMailConfig = ''
                if (configuration.gitUserName && configuration.gitUserEMail)
                    gitUserMailConfig = "-c user.email=\"${configuration.gitUserEMail}\" -c user.name \"${configuration.gitUserName}\""

                try {
                    sh "git ${gitUserMailConfig} commit -m 'update version ${newVersion}'"
                } catch (e) {
                    error "[${STEP_NAME}]git commit failed: ${e}"
                }
                sh "git remote set-url origin ${configuration.gitSshUrl}"
                sh "git tag ${configuration.tagPrefix}${newVersion}"
                sh "git push origin ${configuration.tagPrefix}${newVersion}"

                gitCommitId = gitUtils.getGitCommitIdOrNull()
            }
        }

        if (buildTool == 'docker' && configuration.artifactType == 'appContainer') {
            script.commonPipelineEnvironment.setAppContainerProperty('artifactVersion', newVersion)
            script.commonPipelineEnvironment.setAppContainerProperty('gitCommitId', gitCommitId)
        } else {
            //standard case
            script.commonPipelineEnvironment.setArtifactVersion(newVersion)
            script.commonPipelineEnvironment.setGitCommitId(gitCommitId)
        }

        echo "[${STEP_NAME}]New version: ${newVersion}"
    }
}

def getTimestamp(pattern){
    return sh(returnStdout: true, script: "date --universal +'${pattern}'").trim()
}
