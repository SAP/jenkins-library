import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.MtaUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Utils.downloadSettingsFromUrl

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /** The name of the application which is being built. If the parameter has been provided and no `mta.yaml` exists, the `mta.yaml` will be automatically generated using this parameter and the information (`name` and `version`) from `package.json` before the actual build starts.*/
    'applicationName',
    /**
     * The target platform to which the mtar can be deployed.
     * @possibleValues 'CF', 'NEO', 'XSA'
     */
    'buildTarget',
    /** @see dockerExecute */
    'dockerImage',
    /** The path to the extension descriptor file.*/
    'extension',
    /**
     * The location of the SAP Multitarget Application Archive Builder jar file, including file name and extension.
     * If it is not provided, the SAP Multitarget Application Archive Builder is expected on PATH.
     */
    'mtaJarLocation',
    /** Path or url to the mvn settings file that should be used as global settings file.*/
    'globalSettingsFile',
    /** Path or url to the mvn settings file that should be used as project settings file.*/
    'projectSettingsFile'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** @see dockerExecute */
    'dockerOptions',
    /** Url to the npm registry that should be used for installing npm dependencies.*/
    'defaultNpmRegistry'
])

/**
 * Executes the SAP Multitarget Application Archive Builder to create an mtar archive of the application.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {

            String projectSettingsFile = configuration.projectSettingsFile?.trim()
            if (projectSettingsFile) {
                if (projectSettingsFile.startsWith("http")) {
                    projectSettingsFile = downloadSettingsFromUrl(this, projectSettingsFile, 'project-settings.xml')
                }
                sh 'mkdir -p $HOME/.m2'
                sh "cp ${projectSettingsFile} \$HOME/.m2/settings.xml"
            }

            String globalSettingsFile = configuration.globalSettingsFile?.trim()
            if (globalSettingsFile) {
                if (globalSettingsFile.startsWith("http")) {
                    globalSettingsFile = downloadSettingsFromUrl(this, globalSettingsFile, 'global-settings.xml')
                }
                sh "cp ${globalSettingsFile} \$M2_HOME/conf/settings.xml"
            }

            String defaultNpmRegistry = configuration.defaultNpmRegistry?.trim()
            if (defaultNpmRegistry) {
                sh "npm config set registry $defaultNpmRegistry"
            }

            def mtaYamlName = "mta.yaml"
            def applicationName = configuration.applicationName

            if (!fileExists(mtaYamlName)) {
                if (!applicationName) {
                    error "'${mtaYamlName}' not found in project sources and 'applicationName' not provided as parameter - cannot generate '${mtaYamlName}' file."
                } else {
                    echo "[INFO] '${mtaYamlName}' file not found in project sources, but application name provided as parameter - generating '${mtaYamlName}' file."
                    MtaUtils mtaUtils = new MtaUtils(this)
                    mtaUtils.generateMtaDescriptorFromPackageJson("package.json", mtaYamlName, applicationName)
                }
            } else {
                echo "[INFO] '${mtaYamlName}' file found in project sources."
            }

            def mtaYaml = readYaml file: mtaYamlName

            //[Q]: Why not yaml.dump()? [A]: This reformats the whole file.
            sh "sed -ie \"s/\\\${timestamp}/`date +%Y%m%d%H%M%S`/g\" \"${mtaYamlName}\""

            def id = mtaYaml.ID
            if (!id) {
                error "Property 'ID' not found in ${mtaYamlName} file."
            }

            def mtarFileName = "${id}.mtar"
            // If it is not configured, it is expected on the PATH
            def mtaJar = 'java -jar '
            mtaJar += configuration.mtaJarLocation ?: 'mta.jar'
            def buildTarget = configuration.buildTarget

            def mtaCall = "${mtaJar} --mtar ${mtarFileName} --build-target=${buildTarget}"

            if (configuration.extension) mtaCall += " --extension=$configuration.extension"
            mtaCall += ' build'

            echo "[INFO] Executing mta build call: '${mtaCall}'."

            //[Q]: Why extending the path? [A]: To be sure e.g. grunt can be found
            //[Q]: Why escaping \$PATH ? [A]: We want to extend the PATH variable in e.g. the container and not substituting it with the Jenkins environment when using ${PATH}
            sh """#!/bin/bash
            export PATH=./node_modules/.bin:\$PATH
            $mtaCall
            """

            script?.commonPipelineEnvironment?.setMtarFilePath(mtarFileName)
        }
    }
}
