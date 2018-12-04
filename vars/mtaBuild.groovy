import static com.sap.piper.Prerequisites.checkScript
import static java.lang.Boolean.getBoolean

import com.sap.piper.ConfigurationHelper
import com.sap.piper.MtaUtils
import com.sap.piper.Utils
import com.sap.piper.tools.JavaArchiveDescriptor
import com.sap.piper.tools.ToolDescriptor

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    'applicationName',
    'buildTarget',
    'dockerImage',
    'extension',
    'mtaJarLocation'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'dockerOptions'
])

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

        boolean oldStyleMtaCall = false

        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
            def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
            def mta = new JavaArchiveDescriptor('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '1.0.6', '-v', java)

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
            def mtaCall = 'mtaBuild'

            oldStyleMtaCall = sh(returnStatus: true, script: "which ${mtaCall} > /dev/null") != 0

            if(oldStyleMtaCall) {
                mtaCall = mta.getCall(this, configuration)
            }

            mtaCall += " --mtar ${mtarFileName} --build-target=${configuration.buildTarget}"

            if (configuration.extension) mtaCall += " --extension=$configuration.extension"
            mtaCall += ' build'

            echo "[INFO] Executing mta build call: '${mtaCall}'."

            sh """#!/bin/bash
            export PATH=./node_modules/.bin:${PATH}
            $mtaCall
            """

            script?.commonPipelineEnvironment?.setMtarFilePath(mtarFileName)
        }

        new Utils().pushToSWA([step: STEP_NAME,
                                stepParam1: parameters?.script == null,
                                stepParam2: oldStyleMtaCall], configuration)

        if(oldStyleMtaCall) {

            currentBuild.setResult('UNSTABLE')

            echo "[WARNING] Deprecation warning: An old style mta call has" +
                " been detected. Calling mta this way will be disabled" +
                " in future versions. For more details check the step" +
                " docu at: 'https://sap.github.io/jenkins-library/steps/mtaBuild/'."
        }
    }
}
