import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/onapsisExecuteScan.yaml'
@Field String ONAPSIS_REPORT_NAME = 'onapsis_scan_report.zip'

def call(Map parameters = [:]) {
    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
    } finally {
        def artifact = new File(ONAPSIS_REPORT_NAME)
        if (artifact.exists())
            archiveArtifacts(ONAPSIS_REPORT_NAME)
    }
}
