import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapAddonAssemblyKitCheck.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'abapAddonAssemblyKitCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        [type: 'token', id: 'abapAddonAssemblyKitCertificateFileCredentialsId', env: ['PIPER_abapAddonAssemblyKitCertificateFile']],
        [type: 'token', id: 'abapAddonAssemblyKitCertificatePassCredentialsId', env: ['PIPER_abapAddonAssemblyKitCertificatePass']]
    ]
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
