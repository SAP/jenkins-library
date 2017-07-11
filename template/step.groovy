import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.Utils
/**
 * Name of library step
 *
 * @param script global script environment of the Jenkinsfile run
 * @param others document all parameters
 */
def call(Map parameters = [:], body) {
    //ToDo: Change parameter stepName
    handlePipelineStepErrors (stepName: 'stepName', stepParameters: parameters) {
        def utils = new Utils()
        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]
        //mandatory parameter - default cannot be null
        def mandatoryPara = utils.getMandatoryParameter(parameters, 'paramName', 'param_default')
        //optional parameter - default can be null
        def param1 = parameters.get('param1Name')
    }
}
