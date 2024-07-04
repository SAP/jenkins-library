import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getTransportRequestId
import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'changeManagement',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'changeDocumentLabel',
        /**
        * Defines where the transport request is created, e.g. SAP Solution Manager, ABAP System.
        * @parentConfigKey changeManagement
        * @possibleValues `SOLMAN`, `CTS`, `RFC`
        */
        'type',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'clientOpts',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'credentialsId',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'endpoint',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'git/from',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'git/to',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'git/format',
        /**
         * @see transportRequestCreate
         * @parentConfigKey changeManagement
         */
        'rfc/developmentInstance',
        /**
         * @see transportRequestCreate
         * @parentConfigKey changeManagement
         */
        'rfc/developmentClient',
        /**
         * A pattern used for identifying lines holding the transport request id.
         * @parentConfigKey changeManagement
         */
        'transportRequestLabel',
        /**
         * Some CTS related transport related steps are cm_client based, others are node based.
         * For the node based steps the docker image is specified here.
         * @parentConfigKey changeManagement
         */
        'cts/nodeDocker/image',
        /**
         * The ABAP client. Only for `CTS`
         * @parentConfigKey changeManagement
         */
        'client',
        /**
         * By default we use a standard node docker image and prepare some fiori related packages
         * before performing the deployment. For that we need to launch the image with root privileges.
         * After that, before actually performing the deployment we swith to a non root user. This user
         * can be specified here.
         * @parentConfigKey changeManagement
         */
        'cts/osDeployUser',
        /**
         * By default we use a standard node docker iamge and prepare some fiori related packages
         * performing the deployment. The additional dependencies can be provided here. In case you
         * use an already prepared docker image which contains the required dependencies, the empty
         * list can be provided here. Caused hereby installing additional dependencies will be skipped.
         *
         * @parentConfigKey changeManagement
         */
        'cts/deployToolDependencies',
        /**
         * A list containing additional options for the npm install call. `-g`, `--global` is always assumed.
         * Can be used for e.g. providing custom registries (`--registry https://your.registry.com`) or
         * for providing the verbose flag (`--verbose`) for troubleshooting.
         * @parentConfigKey changeManagement
         */
        'cts/npmInstallOpts',
        /**
         * The file handed over to `fiori deploy` with flag `-c --config`.
         * @parentConfigKey changeManagement
         */
        'cts/deployConfigFile',
  ]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
        /** The name of the application. `RFC` and `CTS` only. */
        'applicationName', // RFC, CTS
        /** The id of the application. Only for `SOLMAN`.*/
        'applicationId', // SOLMAN
        /** The application description, `RFC` and `CTS` only. For `CTS`: the desription is only
            taken into account for a new upload. In case of an update the description will not be
            updated.
        */
        'applicationDescription',
        /** The path of the file to upload, Only for `SOLMAN`.*/
        'filePath', // SOLMAN
        /** The URL where to find the UI5 package to upload to the transport request.  Only for `RFC`. */
        'applicationUrl', // RFC
        /** The ABAP package name of your application. */
        'abapPackage',
        /** The code page of your ABAP system. E.g. UTF-8. */
        'codePage', //RFC
        /** If unix style line endings should be accepted. Only for `RFC`.*/
        'acceptUnixStyleLineEndings', // RFC
        /** @see transportRequestCreate */
        'verbose', // RFC
    ])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** @see transportRequestCreate */
    'changeDocumentId',
    /** The id of the transport request to upload the file. This parameter is only taken into account
      * when provided via signature to the step.
      */
    'transportRequestId'])

/** Uploads a file to a Transport Request. */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('filePath', script.commonPipelineEnvironment.getMtarFilePath())

        Map configuration = configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
            .collectValidationFailures()
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/client', null, {backendType == BackendType.CTS})
            .withMandatoryProperty('changeManagement/type')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('filePath', null, { backendType == BackendType.SOLMAN })
            .withMandatoryProperty('applicationUrl', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('codePage', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('acceptUnixStyleLineEndings', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('changeManagement/rfc/developmentInstance', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('changeManagement/rfc/developmentClient', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('changeManagement/rfc/docker/image', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/docker/options', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/docker/envVars', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/docker/pullImage', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('applicationDescription', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('abapPackage', null, { backendType in [BackendType.RFC, BackendType.CTS] })
            .withMandatoryProperty('applicationId', null, {backendType == BackendType.SOLMAN})
            .withMandatoryProperty('applicationName', null, {backendType in [BackendType.RFC, BackendType.CTS]})
            .withMandatoryProperty('failOnWarning', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('verbose', null, {backendType == BackendType.RFC})

        def changeDocumentId = null

        if(backendType == BackendType.SOLMAN) {
            changeDocumentId = getChangeDocumentId(script, configuration)
        }

        def transportRequestId = getTransportRequestId(script, configuration)

        configHelper
            .mixin([changeDocumentId: changeDocumentId?.trim() ?: null,
                    transportRequestId: transportRequestId?.trim() ?: null], ['changeDocumentId', 'transportRequestId'] as Set)

        if(backendType == BackendType.SOLMAN) {
            configHelper
                .withMandatoryProperty('changeDocumentId',
                    "Change document id not provided (parameter: \'changeDocumentId\' provided to the step call or via commit history).")
        }
        configuration = configHelper
            .withMandatoryProperty('transportRequestId',
                "Transport request id not provided (parameter: \'transportRequestId\' provided to the step call or via commit history).")
            .use()

            try {

                switch(backendType) {

                    case BackendType.SOLMAN:

                        echo "[INFO] Uploading file '${configuration.filePath}' to transport request '${configuration.transportRequestId}'" +
                            " of change document '${configuration.changeDocumentId}'."

                        Map paramsUpload = [
                            script: script,
                            filePath: configuration.filePath,
                            uploadCredentialsId: configuration.changeManagement.credentialsId,
                            endpoint: configuration.changeManagement.endpoint,
                            applicationId: configuration.applicationId,
                            changeDocumentId: configuration.changeDocumentId,
                            transportRequestId: configuration.transportRequestId
                            ]

                        if(configuration.changeManagement.clientOpts) {
                            paramsUpload.cmClientOpts = configuration.changeManagement.clientOpts
                        }
                        paramsUpload = addDockerParams(script, paramsUpload, configuration.changeManagement.solman?.docker)

                        transportRequestUploadSOLMAN(paramsUpload)

                        echo "[INFO] File '${configuration.filePath}' has been successfully uploaded to transport request '${configuration.transportRequestId}'" +
                            " of change document '${configuration.changeDocumentId}'."

                        break
                    case BackendType.CTS:

                        echo "[INFO] Uploading application '${configuration.applicationName}' to transport request '${configuration.transportRequestId}'."

                        Map paramsUpload = [
                            script: script,
                            transportRequestId: configuration.transportRequestId,
                            endpoint: configuration.changeManagement.endpoint,
                            client: configuration.changeManagement.client,
                            applicationName: configuration.applicationName,
                            description: configuration.applicationDescription,
                            abapPackage: configuration.abapPackage,
                            osDeployUser: configuration.changeManagement.cts.osDeployUser,
                            deployToolDependencies: configuration.changeManagement.cts.deployToolDependencies,
                            npmInstallOpts: configuration.changeManagement.cts.npmInstallOpts,
                            deployConfigFile: configuration.changeManagement.cts.deployConfigFile,
                            uploadCredentialsId: configuration.changeManagement.credentialsId
                            ]

                        paramsUpload = addDockerParams(script, paramsUpload, configuration.changeManagement.cts?.nodeDocker)

                        transportRequestUploadCTS(paramsUpload)

                        echo "[INFO] Application '${configuration.applicationName}' has been successfully uploaded to transport request '${configuration.transportRequestId}'."

                        break
                    case BackendType.RFC:

                        echo "[INFO] Uploading file '${configuration.applicationUrl}' to transport request '${configuration.transportRequestId}'."

                        Map paramsUpload = [script: script,
                            transportRequestId: configuration.transportRequestId,
                            applicationName: configuration.applicationName,
                            applicationUrl: configuration.applicationUrl,
                            endpoint: configuration.changeManagement.endpoint,
                            uploadCredentialsId: configuration.changeManagement.credentialsId,
                            instance: configuration.changeManagement.rfc.developmentInstance,
                            client: configuration.changeManagement.rfc.developmentClient,
                            applicationDescription: configuration.applicationDescription,
                            abapPackage: configuration.abapPackage,
                            codePage: configuration.codePage,
                            acceptUnixStyleLineEndings: configuration.acceptUnixStyleLineEndings,
                            failUploadOnWarning: configuration.failOnWarning,
                            verbose: configuration.verbose
                            ]

                        paramsUpload = addDockerParams(script, paramsUpload, configuration.changeManagement.rfc?.docker)

                        transportRequestUploadRFC(paramsUpload)

                        echo "[INFO] File 'configuration.applicationUrl' has been successfully uploaded to transport request '${configuration.transportRequestId}'."

                        break
                }

            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
    }
}

Map addDockerParams(Script script, Map parameters, Map docker) {

    if(docker) {
        if(docker.image) {
            parameters.dockerImage = docker.image
        }
        if(docker.options) {
            parameters.dockerOptions = docker.options
        }
        if(docker.envVars) {
            parameters.dockerEnvVars = docker.envVars
        }
        if(docker.pullImage != null) {
            parameters.put('dockerPullImage', docker.pullImage)
        }
    }
    return parameters
}
