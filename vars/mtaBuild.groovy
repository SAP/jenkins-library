import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {

    piperExecuteBin parameters, STEP_NAME, "metadata/${STEP_NAME}.yaml", []
}
