package com.sap.piper

class TemporaryCredentialsUtils implements Serializable {

    private static Script script

    TemporaryCredentialsUtils(Script script) {
        this.script = script
    }

    void handleTemporaryCredentials(List credentialItems, String credentialsDirectory, Closure body) {
        final String credentialsFileName = 'credentials.json'

        if (credentialsDirectory == null) {
            script.error("This should not happen: Directory for credentials file not specified.")
        }

        try {
            if (credentialItems != null && !credentialItems.isEmpty()) {
                writeCredentials(credentialItems, credentialsDirectory, credentialsFileName)
            }
            body()
        }
        finally {
            if (credentialItems != null && !credentialItems.isEmpty()) {
                deleteCredentials(credentialsDirectory, credentialsFileName)
            }
        }
    }

    private void writeCredentials(List credentialItems, String credentialsDirectory, String credentialsFileName) {
        if (credentialItems == null || credentialItems.isEmpty()) {
            script.echo "Not writing any credentials."
            return
        }

        assertSystemsFileExists(credentialsDirectory)

        String credentialJson = readCredentials(credentialItems)

        script.echo "Writing credential file with ${credentialItems.size()} items."
        script.dir(credentialsDirectory) {
            script.writeFile file: credentialsFileName, text: credentialJson
        }
    }

    private void deleteCredentials(String credentialsDirectory, String credentialsFileName) {
        script.echo "Deleting credential file."
        script.dir(credentialsDirectory) {
            script.sh "rm -f ${credentialsFileName}"
        }
    }

    private String readCredentials(List credentialItems) {
        Map credentialCollection = [:]
        credentialCollection['credentials'] = []

        for (int i = 0; i < credentialItems.size(); i++) {
            String alias = credentialItems[i]['alias']
            String jenkinsCredentialId = credentialItems[i]['credentialId']

            script.withCredentials([
                script.usernamePassword(credentialsId: jenkinsCredentialId, passwordVariable: 'password', usernameVariable: 'user')
            ]) {
                credentialCollection['credentials'] += [alias: alias, username: script.user, password: script.password]
            }
        }
        return new JsonUtils().groovyObjectToJsonString(credentialCollection)
    }

    private assertSystemsFileExists(String credentialsDirectory){
        script.dir(credentialsDirectory) {
            if (!script.fileExists("systems.yml") && !script.fileExists("systems.yaml") && !script.fileExists("systems.json")) {
                script.error("The directory ${credentialsDirectory} does not contain any of the files systems.yml, systems.yaml or systems.json. " +
                    "One of those files is required in order to activate the integration test credentials configured in the pipeline configuration file of this project. " +
                    "Please add the file as explained in the SAP Cloud SDK documentation.")
            }
        }
    }
}
