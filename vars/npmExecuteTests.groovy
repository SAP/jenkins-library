import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteTests.yaml'

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = []

void call(Map parameters = [:]) {
    List credentials = []
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
