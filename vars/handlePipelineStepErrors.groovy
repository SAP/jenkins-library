def call(Map parameters = [:], body) {

    def stepParameters = parameters.stepParameters //mandatory
    def stepName = parameters.stepName //mandatory
    def echoDetails = parameters.get('echoDetails', true)

    try {

        if (stepParameters == null && stepName == null)
            error "step handlePipelineStepErrors requires following mandatory parameters: stepParameters, stepName"

        if (echoDetails)
            echo "--- BEGIN LIBRARY STEP: ${stepName}.groovy ---"

        body()

    } catch (Throwable err) {
        if (echoDetails)
            echo """----------------------------------------------------------
--- ERROR OCCURED IN LIBRARY STEP: ${stepName}
----------------------------------------------------------

FOLLOWING PARAMETERS WERE AVAILABLE TO THIS STEP:
***
${stepParameters}
***

ERROR WAS:
***
${err}
***

FURTHER INFORMATION:
* Documentation of library step ${stepName}: https://sap.github.io/jenkins-library/steps/${stepName}/
* Source code of library step ${stepName}: https://github.com/SAP/jenkins-library/blob/master/vars/${stepName}.groovy
* Library documentation: https://sap.github.io/jenkins-library/
* Library repository: https://github.com/SAP/jenkins-library
 
----------------------------------------------------------"""
        throw err
    } finally {
        if (echoDetails)
            echo "--- END LIBRARY STEP: ${stepName}.groovy ---"
    }
}
