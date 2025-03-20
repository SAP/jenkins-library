import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/onapsisExecuteScan.yaml'
@Field String ONAPSIS_REPORT_NAME = 'onapsis_scan_report.zip'

def call(Map parameters = [:]) {

    List credentials = [
        [type: 'token', id: 'onapsisSecretTokenId', env: ['PIPER_onapsisSecretToken'], resolveCredentialsId: false],
        [type: 'file', id: 'onapsisCertificate', env: ['PIPER_onapsisCertificatePath'], resolveCredentialsId: false]
    ]

    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    } finally {
        if (fileExists(ONAPSIS_REPORT_NAME)) {
            archiveArtifacts(ONAPSIS_REPORT_NAME)
        }
    }
}
