import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/onapsisExecuteScan.yaml'
@Field String ONAPSIS_REPORT_NAME = 'onapsis_scan_report.zip'

def call(Map parameters = [:]) {

    List credentials = [
        [type: 'secret', id: 'onapsisSecretTokenId', env: ['PIPER_onapsisSecretToken']],
        [type: 'secret', id: 'onapsisCertificate', env: ['PIPER_onapsisCertificate']]
    ]

    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    } finally {
        def artifact = new File(ONAPSIS_REPORT_NAME)
        if (artifact.exists())
            archiveArtifacts(ONAPSIS_REPORT_NAME)
    }
}
