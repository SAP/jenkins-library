package com.sap.piper

import hudson.AbortException

/**
 * Retrieves the credentialsId either from the job parameters or from the configured SCMs. In case a
 * credentialsId if found in both places the credentialsId from the job parameters is returned.
 * @return The credentialsId
 */
def getCredentialsId(repoUrl, filterPattern = ~'.*', ignorePattern= ~'^- none -$', String key='CREDENTIALS_ID') {

    def credentialsIdFromJobParameters, credentialsIdFromSCMConfig

    try {
        credentialsIdFromSCMConfig = getCredentialsIdFromJobSCMConfig(repoUrl)
    } catch(AbortException ex) {
        // Most probably pipeline script is inlined in the job configuration in this case. In this case there
        // is no SCM configuration in the job.
        echo "[INFO] ${ex.getClass().getName()} caught while retrieving credentialsId from SCM configuration. " +
             "This does not indicate any problem in case the pipeline script is inlined in the job configuration."
    }

    credentialsIdFromJobParameters = getCredentialsIdFromJobParameters(filterPattern, ignorePattern, key)

    if(credentialsIdFromJobParameters) {
        if(credentialsIdFromSCMConfig && credentialsIdFromSCMConfig != credentialsIdFromJobParameters) {
            echo "[WARNING] CredentialsId in SCM configuration ('${credentialsIdFromSCMConfig}') differs from " +
                 "credentialsId retrieved from job parameters ('${credentialsIdFromJobParameters}'). " +
                 "CredentialsId from job parameters will be used."
        }
        echo "[INFO] Returning credentialsId found inside job parameters: '${credentialsIdFromJobParameters}'."
        return credentialsIdFromJobParameters
    }

    if(credentialsIdFromSCMConfig) {
        echo "[INFO] Returning credentialsId found inside SCM configuration: '${credentialsIdFromSCMConfig}'."
    } else {
        echo "[INFO] No credentialsId found, neither in job parameters nor in jobs SCM configuration."
    }

    return credentialsIdFromSCMConfig
}

/**
 * Retrieves the credentialsId from SCMs configured in the job.
 * @return The credentialsId. In case there are no credentials maintained for the repo or in case
 *         the repository is not configured in the job definition <code>null</code> is returned.
 * @param repoUrl. The repository url for that the credentials should be retrieved.
 * @throws IllegalArgumentExcpetion In case <code>repoUrl</code> is null or empty.
 */
def getCredentialsIdFromJobSCMConfig(repoUrl) {

    if(!repoUrl) throw new IllegalArgumentException('repoUrl was null or empty.')

    def credentialsId = retrieveScm().userRemoteConfigs.find( { it -> it.url == repoUrl} )?.getCredentialsId()

    if(!credentialsId) {
        echo "[INFO] No credentialsId found in SCM configuration for repository '${repoUrl}'."
    } else {
        echo "[INFO] CredentialsId '${credentialsId}' found in SCM configuration for repository '${repoUrl}'."
    }

    return credentialsId
}

/**
 * Returns the credentialsId from the job parameters.
 * @param filterPattern A regular expession. Returns the first matching group. Useful in case the value is mixed up with
 *                      other information, like it could happen when using a select box from the extensible choise parameter
 *                      plugin. Default '.*'. With the default the parameter value is returned &quot;as it&quot;.
 * @param ignorePattern A regular expession. If the regex matches, <code>null</code> is returned instead of the value.
 * @param key The key of the parameter holding the credentialsId. Default &quot;CREDENTIALS_ID&quot;
 */
def getCredentialsIdFromJobParameters(filterPattern = ~'.*', ignorePattern = null, key='CREDENTIALS_ID') {

    def credentialsId

    if(params."${key}") {
        if (ignorePattern && params."${key}" ==~ ignorePattern) {
            echo "[INFO] Parameter '${key}' found in job parameters. Value: '${params."${key}"}'. This value will be ignored."
        } else {
            echo "[INFO] Parameter '${key}' found in job parameters. Value: '${params."${key}"}'. Applying filter pattern '${filterPattern}' for extracting credentials id."
            def valueMatcher = (params."${key}" =~ filterPattern)
            credentialsId = valueMatcher.size() > 0 ? valueMatcher[0] : params."${key}"
        }
    } else {
        echo "[INFO] Parameter '${key}' not found in job parameters '${params}'."
    }

    return credentialsId
}

def retrieveScm() {
    return scm
}

