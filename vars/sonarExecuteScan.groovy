import com.sap.piper.GenerateDocumentation

import groovy.transform.Field

import java.nio.charset.StandardCharsets

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sonar.yaml'

    /**
     * The name of the SonarQube instance defined in the Jenkins settings.
     */
//    'instance',

/**
 * The step executes the [sonar-scanner](https://docs.sonarqube.org/display/SCAN/Analyzing+with+SonarQube+Scanner) cli command to scan the defined sources and publish the results to a SonarQube instance.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        //.addIfEmpty('projectVersion', script.commonPipelineEnvironment.getArtifactVersion()?.tokenize('.')?.get(0))
        //.addIfEmpty('githubOrg', script.commonPipelineEnvironment.getGithubOrg())
        //.addIfEmpty('githubRepo', script.commonPipelineEnvironment.getGithubRepo())
        // check mandatory parameters
        //.withMandatoryProperty('githubTokenCredentialsId', null, { config -> config.legacyPRHandling && isPullRequest() })
        //.withMandatoryProperty('githubOrg', null, { isPullRequest() })
        //.withMandatoryProperty('githubRepo', null, { isPullRequest() })

        /*
        if(!script.fileExists('.git')) {
            utils.unstash('git')
        }
        */

        List credentials = [
            [type: 'token', id: 'sonarTokenCredentialsId', env: ['SONAR_TOKEN']],
            [type: 'token', id: 'githubTokenCredentialsId', env: ['GITHUB_TOKEN']]
        ]

        loadCertificates([
            customTlsCertificateLinks: parameters.certificates,
            verbose: false
        ])

        withSonarQubeEnv(parameters.instance) {
            piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
        }
    }
}

private void loadCertificates(Map config) {
    String certificateFolder = '.certificates/'
    List wgetOptions = [
        "--directory-prefix ${certificateFolder}"
    ]
    List keytoolOptions = [
        '-import',
        '-noprompt',
        '-storepass changeit',
        "-keystore ${certificateFolder}cacerts"
    ]
    if (config.customTlsCertificateLinks){
        if(config.verbose){
            wgetOptions.push('--verbose')
            keytoolOptions.push('-v')
        }else{
            wgetOptions.push('--no-verbose')
        }
        config.customTlsCertificateLinks.each { url ->
            def filename = new File(url).getName()
            filename = URLDecoder.decode(filename, StandardCharsets.UTF_8.name())
            sh "wget ${wgetOptions.join(' ')} ${url}"
            sh "keytool ${keytoolOptions.join(' ')} -alias '${filename}' -file '${certificateFolder}${filename}'"
        }
    }
}
