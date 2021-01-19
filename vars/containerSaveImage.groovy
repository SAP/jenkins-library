import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/savecontainer.yaml'

//Metadata maintained in file project://resources/metadata/savecontainer.yaml

void call(Map parameters = [:]) {
    List credentials = []
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
