import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/onapsisExecuteScan.yaml'

def call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'onapsisTokenCredentialsId', env: ['PIPER_accessToken']]
        // [type: 'string', id: 'scanServiceUrl', env: ['PIPER_scanServiceUrl']],
        // [type: 'string', id: 'onapsisUsername', env: ['PIPER_onapsisUsername']],
        // [type: 'secret', id: 'onapsisPassword', env: ['PIPER_onapsisPassword']]
    ]

    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)

    } catch (Exception e) {
        error("An error occurred while executing the Onapsis scan: ${e.message}")
        currentBuild.result = 'FAILURE' // Mark the build as failed
        throw e // Stop execution and fail the build immediately
    } finally {
        archiveArtifacts('onapsis_scan_report.zip')
    }
}
