import com.sap.piper.ConfigurationMerger

def call(Map parameters = [:]) {

    handlePipelineStepErrors(stepName: 'mavenExecute', stepParameters: parameters) {
        final script = parameters.script

        prepareDefaultValues script: script

        Set parameterKeys = [
            'dockerImage',
            'dockerOptions',
            'globalSettingsFile',
            'projectSettingsFile',
            'pomPath',
            'flags',
            'goals',
            'm2Path',
            'defines',
            'logSuccessfulMavenTransfers'
        ]
        Set stepConfigurationKeys = [
            'dockerImage',
            'globalSettingsFile',
            'projectSettingsFile',
            'pomPath',
            'm2Path'
        ]

        Map configuration = ConfigurationMerger.merge(script, 'mavenExecute',
                                                      parameters, parameterKeys,
                                                      stepConfigurationKeys)

        String command = "mvn"

        def globalSettingsFile = configuration.globalSettingsFile
        if (globalSettingsFile?.trim()) {
            if(globalSettingsFile.trim().startsWith("http")){
                downloadSettingsFromUrl(globalSettingsFile)
                globalSettingsFile = "settings.xml"
            }
            command += " --global-settings '${globalSettingsFile}'"
        }

        def m2Path = configuration.m2Path
        if(m2Path?.trim()) {
            command += " -Dmaven.repo.local='${m2Path}'"
        }

        def projectSettingsFile = configuration.projectSettingsFile
        if (projectSettingsFile?.trim()) {
            if(projectSettingsFile.trim().startsWith("http")){
                downloadSettingsFromUrl(projectSettingsFile)
                projectSettingsFile = "settings.xml"
            }
            command += " --settings '${projectSettingsFile}'"
        }

        def pomPath = configuration.pomPath
        if(pomPath?.trim()){
            command += " --file '${pomPath}'"
        }

        def mavenFlags = configuration.flags
        if (mavenFlags?.trim()) {
            command += " ${mavenFlags}"
        }

        // Always use Maven's batch mode
        if (!(command.contains('-B') || command.contains('--batch-mode'))){
            command += ' --batch-mode'
        }

        // Disable log for successful transfers by default. Note this requires the batch-mode flag.
        final String disableSuccessfulMavenTransfersLogFlag = ' -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn'
        if (!configuration.logSuccessfulMavenTransfers) {
            if (!command.contains(disableSuccessfulMavenTransfersLogFlag)) {
                command += disableSuccessfulMavenTransfersLogFlag
            }
        }

        def mavenGoals = configuration.goals
        if (mavenGoals?.trim()) {
            command += " ${mavenGoals}"
        }
        def defines = configuration.defines
        if (defines?.trim()){
            command += " ${defines}"
        }
        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
            sh command
        }
    }
}

private downloadSettingsFromUrl(String url){
    def settings = httpRequest url
    writeFile file: 'settings.xml', text: settings.getContent()

}

