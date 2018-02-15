import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger

def call(Map parameters = [:]) {

    handlePipelineStepErrors(stepName: 'mavenExecute', stepParameters: parameters) {
        final script = parameters.script

        prepareDefaultValues script: script
        final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, 'mavenExecute')

        final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, 'mavenExecute')

        List parameterKeys = [
            'dockerImage',
            'globalSettingsFile',
            'projectSettingsFile',
            'pomPath',
            'flags',
            'goals',
            'm2Path',
            'defines'
        ]
        List stepConfigurationKeys = [
            'dockerImage',
            'globalSettingsFile',
            'projectSettingsFile',
            'pomPath',
            'm2Path'
        ]

        Map configuration = ConfigurationMerger.merge(parameters, parameterKeys, stepConfiguration, stepConfigurationKeys, stepDefaults)

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

        def mavenGoals = configuration.goals
        if (mavenGoals?.trim()) {
            command += " ${mavenGoals}"
        }
        def defines = configuration.defines
        if (defines?.trim()){
            command += " ${defines}"
        }

        dockerExecute(dockerImage: configuration.dockerImage) { sh command }
    }
}

private downloadSettingsFromUrl(String url){
    def settings = httpRequest url
    writeFile file: 'settings.xml', text: settings.getContent()

}

