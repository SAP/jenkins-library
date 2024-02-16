import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/emitEvents.yaml'

void call(Map parameters = [:]) {
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
}