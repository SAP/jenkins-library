package com.sap.piper

class TemporaryCredentialsUtils implements Serializable {

    private Script script

    TemporaryCredentialsUtils(Script script) {
        this.script = script
    }

    void handleTemporaryCredentials(List credentialItems, List credentialsDirectories, Closure body) {
        final String credentialsFileName = 'credentials.json'

        if (!credentialsDirectories) {
            script.error("This should not happen: Directories for credentials files not specified.")
        }

        final boolean useCredentials = credentialItems
        try {
            if (useCredentials) {
                writeCredentials(credentialItems, credentialsDirectories, credentialsFileName)
            }
            body()
        }
        finally {
            if (useCredentials) {
                deleteCredentials(credentialsDirectories, credentialsFileName)
            }
        }
    }

    private void writeCredentials(List credentialItems, List credentialsDirectories, String credentialsFileName) {
        if (!credentialItems) {
            script.echo "Not writing any credentials."
            return
        }

        Boolean systemsFileFound = false
        for (int i = 0; i < credentialsDirectories.size(); i++) {
            if (!credentialsDirectories[i]) {
                continue
            }
            if (!credentialsDirectories[i].endsWith("/")) {
                credentialsDirectories[i] += '/'
            }
            if (script.fileExists("${credentialsDirectories[i]}systems.yml") || script.fileExists("${credentialsDirectories[i]}systems.yaml") || script.fileExists("${credentialsDirectories[i]}systems.json")) {
                String credentialJson = returnCredentialsAsJSON(credentialItems)

                script.echo "Writing credentials file with ${credentialItems.size()} items to ${credentialsDirectories[i]}."
                script.writeFile file: credentialsDirectories[i] + credentialsFileName, text: credentialJson

                systemsFileFound = true
            }
        }
        if (!systemsFileFound) {
            script.error("None of the directories ${credentialsDirectories} contains any of the files systems.yml, systems.yaml or systems.json. " +
                "One of those files is required in order to activate the integration test credentials configured in the pipeline configuration file of this project. " +
                "Please add the file as explained in project 'Piper' documentation.")
        }
    }

    private void deleteCredentials(List credentialsDirectories, String credentialsFileName) {
        for (int i = 0; i < credentialsDirectories.size(); i++) {
            if (!credentialsDirectories[i]) {
                continue
            }
            if(!credentialsDirectories[i].endsWith('/'))
                credentialsDirectories[i] += '/'

            if (script.fileExists(credentialsDirectories[i] + credentialsFileName)) {
                script.echo "Deleting credentials file in ${credentialsDirectories[i]}."
                script.sh "rm -f ${credentialsDirectories[i] + credentialsFileName}"
            }
        }
    }

    private String returnCredentialsAsJSON(List credentialItems) {
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
}
