import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentReadAddonDescriptor.yaml'

void call(Map parameters = [:]) {
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, null, false, false, true)
}
