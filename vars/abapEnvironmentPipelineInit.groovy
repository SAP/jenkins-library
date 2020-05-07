import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

/**
 * This stage initializes the ABAP Environment Pipeline run
 */
@GenerateStageDocumentation(defaultStageName = 'Init')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this

    setupCommonPipelineEnvironment script: script

    echo "Config: ${script.commonPipelineEnvironment.configuration}"
    echo "Config File: ${script.commonPipelineEnvironment.configurationFile}"

}
