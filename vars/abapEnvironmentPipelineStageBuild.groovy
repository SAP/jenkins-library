import groovy.transform.Field
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    'cloudFoundryCreateServiceKey',
    'abapAddonAssemblyKitReserveNextPackages',
    'abapEnvironmentAssemblePackages',
    'abapAddonAssemblyKitRegisterPackages',
    'abapAddonAssemblyKitReleasePackages',
    'abapEnvironmentAssembleConfirm',
    'abapAddonAssemblyKitCreateTargetVector',
    'abapAddonAssemblyKitPublishTargetVector'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage builds an AddOn for the SAP Cloud Platform ABAP Environment
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        cloudFoundryCreateServiceKey script: parameters.script
        abapEnvironmentAssemblePackages script: parameters.script
        abapAddonAssemblyKitRegisterPackages script: parameters.script
        abapAddonAssemblyKitReleasePackages script: parameters.script
        abapEnvironmentAssembleConfirm script: parameters.script
        //abapAddonAssemblyKitCreateTargetVector script: parameters.script
        //abapAddonAssemblyKitPublishTargetVector(script: parameters.script, targetVectorScope: 'T')
    }

}
