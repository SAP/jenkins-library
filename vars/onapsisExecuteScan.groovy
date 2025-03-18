import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/onapsisExecuteScan.yaml'

def call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        try {
            piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
        } finally {
            archiveArtifacts('onapsis_scan_report.zip')
        }
    }
}
