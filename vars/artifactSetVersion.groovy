import com.sap.piper.ConfigurationMerger
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning
import groovy.text.SimpleTemplateEngine

def call(Map parameters = [:]) {

    def stepName = 'artifactSetVersion'

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

        def gitUtils = parameters.juStabGitUtils
        if (gitUtils == null) {
            gitUtils = new GitUtils()
        }

        if (fileExists('.git')) {
            if (sh(returnStatus: true, script: 'git diff --quiet HEAD') != 0)
                error "[${stepName}] Files in the workspace have been changed previously - aborting ${stepName}"
        }

        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        prepareDefaultValues script: script

        Set parameterKeys = [
            'artifactType',
            'buildTool',
            'commitVersion',
            'dockerVersionSource',
            'filePath',
            'gitCommitId',
            'gitCredentialsId',
            'gitUserEMail',
            'gitUserName',
            'gitSshUrl',
            'tagPrefix',
            'timestamp',
            'timestampTemplate',
            'versioningTemplate'
        ]
        Map pipelineDataMap = [
            gitCommitId: gitUtils.getGitCommitIdOrNull()
        ]
        Set stepConfigurationKeys = [
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

        Map configuration = ConfigurationMerger.merge(script, stepName, parameters, parameterKeys, pipelineDataMap, stepConfigurationKeys)

        def utils = new Utils()
        def buildTool = utils.getMandatoryParameter(configuration, 'buildTool')

        if (!configuration.filePath)
            configuration.filePath = configuration[buildTool].filePath //use default configuration

        def newVersion
        def artifactVersioning = ArtifactVersioning.getArtifactVersioning(buildTool, this, configuration)

        if(configuration.artifactType == 'appContainer' && configuration.dockerVersionSource == 'appVersion'){
            if (script.commonPipelineEnvironment.getArtifactVersion())
                //replace + sign if available since + is not allowed in a Docker tag
                newVersion = script.commonPipelineEnvironment.getArtifactVersion().replace('+', '_')
            else
                error ("[${stepName}] No artifact version available for 'dockerVersionSource: appVersion' -> executeBuild needs to run for the application artifact first to set the artifactVersion for the application artifact.'")
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
                    error "[${stepName}]git commit failed: ${e}"
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

        echo "[${stepName}]New version: ${newVersion}"
    }
}

def getTimestamp(pattern){
    return sh(returnStdout: true, script: "date +'${pattern}'").trim()
}





