import groovy.text.GStringTemplateEngine

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /**
     * Defines the build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.
     */
    'containerBuildOptions',
    /** @see dockerExecute */
    'containerCommand',
    /** Defines the full name of the Docker image to be created including registry, image name and tag like `my.docker.registry/path/myImageName:myTag`.*/
    'containerImageNameAndTag',
    /** @see dockerExecute */
    'containerShell',
    /**
     * Defines the command to prepare the Kaniko container.
     * By default the contained credentials are removed in order to allow anonymous access to container registries.
     */
    'containerPreparationCommand',
    /**
     * List containing download links of custom TLS certificates. This is required to ensure trusted connections to registries with custom certificates.
     */
    'customTlsCertificateLinks',
    /**
     * Defines the location of the Dockerfile relative to the Jenkins workspace.
     */
    'dockerfile',
    /**
     * Defines the id of the file credentials in your Jenkins credentials store which contain the file `.docker/config.json`.
     * You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).
     */
    'dockerConfigJsonCredentialsId',
    /** @see dockerExecute */
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute */
    'dockerImage'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // telemetry reporting
        new Utils().pushToSWA([
            step: STEP_NAME
        ], config)

        def buildOptions = new GStringTemplateEngine().createTemplate(config.containerBuildOptions).make([config: config, env: env]).toString()

        if (!buildOptions.contains('--destination')) {
            if (config.containerImageNameAndTag) {
                buildOptions += " --destination ${config.containerImageNameAndTag}"
            } else {
                buildOptions += " --no-push"
            }
        }

        dockerExecute(
            script: script,
            containerCommand: config.containerCommand,
            containerShell: config.containerShell,
            dockerEnvVars: config.dockerEnvVars,
            dockerImage: config.dockerImage,
            dockerOptions: config.dockerOptions
        ) {
            // prepare kaniko container for running with proper Docker config.json and custom certificates
            // custom certificates will be downloaded and appended to ca-certificates.crt file used in container
            sh """#!${config.containerShell}
${config.containerPreparationCommand}
${getCertificateUpdate(config.customTlsCertificateLinks)}
"""

            def uuid = UUID.randomUUID().toString()
            if (config.dockerConfigJsonCredentialsId) {
                // write proper config.json with credentials
                withCredentials([file(credentialsId: config.dockerConfigJsonCredentialsId, variable: 'dockerConfigJson')]) {
                    writeFile file: "${uuid}-config.json", text: readFile(dockerConfigJson)
                }
            } else {
                // empty config.json to allow anonymous authentication
                writeFile file: "${uuid}-config.json", text: '{"auths":{}}'
            }

            // execute Kaniko
            sh """#!${config.containerShell}
mv ${uuid}-config.json /kaniko/.docker/config.json
/kaniko/executor --dockerfile ${env.WORKSPACE}/${config.dockerfile} --context ${env.WORKSPACE} ${buildOptions}"""
        }
    }
}

private String getCertificateUpdate(List certLinks) {
    String certUpdate = ''

    if (!certLinks) return certUpdate

    certLinks.each {link ->
        certUpdate += "wget ${link} -O - >> /kaniko/ssl/certs/ca-certificates.crt\n"
    }
    return certUpdate
}
