import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/onapsisExecuteScan.yaml'

def call(Map parameters = [:]) {

    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
    } catch (Exception e) {
//        error("An error occurred while executing the Onapsis scan: ${e.message}")
        currentBuild.result = 'FAILURE' // Mark the build as failed
    } finally {
        archiveArtifacts('onapsis_scan_report.zip')
    }
}
