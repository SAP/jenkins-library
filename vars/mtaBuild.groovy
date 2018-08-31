import com.sap.piper.ConfigurationHelper
import com.sap.piper.MtaUtils
import com.sap.piper.Utils
import com.sap.piper.tools.JavaArchiveDescriptor
import com.sap.piper.tools.ToolDescriptor

import groovy.transform.Field

@Field def STEP_NAME = 'mtaBuild'

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

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        // load default & individual configuration
        Map configuration = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        def mtarFileName = ""

        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
            def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
            java.verify(this, configuration)

            def mta = new JavaArchiveDescriptor('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '1.0.6', '-v', java)
            mta.verify(this, configuration)

            def mtaYmlName = "${pwd()}/mta.yaml"
            def applicationName = configuration.applicationName

            if (!fileExists(mtaYmlName)) {
                if (!applicationName) {
                    echo "'applicationName' not provided as parameter - will not try to generate mta.yml file"
                } else {
                    MtaUtils mtaUtils = new MtaUtils(this)
                    mtaUtils.generateMtaDescriptorFromPackageJson("${pwd()}/package.json", mtaYmlName, applicationName)
                }
            }

            def mtaYaml = readYaml file: "${pwd()}/mta.yaml"

            //[Q]: Why not yaml.dump()? [A]: This reformats the whole file.
            sh "sed -ie \"s/\\\${timestamp}/`date +%Y%m%d%H%M%S`/g\" \"${pwd()}/mta.yaml\""

            def id = mtaYaml.ID
            if (!id) {
                error "Property 'ID' not found in mta.yaml file at: '${pwd()}'"
            }

            mtarFileName = "${id}.mtar"
            def mtaJar = mta.getCall(this, configuration)
            def buildTarget = configuration.buildTarget

            def mtaCall = "${mtaJar} --mtar ${mtarFileName} --build-target=${buildTarget}"

            if (configuration.extension) mtaCall += " --extension=$configuration.extension"
            mtaCall += ' build'

            sh """#!/bin/bash
                export PATH=./node_modules/.bin:${PATH}
                $mtaCall
                """
        }

        def mtarFilePath = "${pwd()}/${mtarFileName}"
        script?.commonPipelineEnvironment?.setMtarFilePath(mtarFilePath)
    }
}

