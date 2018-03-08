import com.sap.piper.ConfigurationMerger
import com.sap.piper.MtaUtils
import com.sap.piper.tools.Tool
import com.sap.piper.tools.ToolVerifier
import com.sap.piper.tools.ToolUtils

import groovy.transform.Field


def call(Map parameters = [:]) {

    def stepName = 'mtaBuild'

    Set parameterKeys = [
        'applicationName',
        'buildTarget',
        'mtaJarLocation'
    ]

    Set stepConfigurationKeys = [
        'applicationName',
        'buildTarget',
        'mtaJarLocation'
    ]

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        prepareDefaultValues script: script

        final Map configuration = ConfigurationMerger.merge(
                                      script, stepName,
                                      parameters, parameterKeys,
                                      stepConfigurationKeys)

        def mta = new Tool('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v')
        ToolVerifier.verifyToolVersion(mta, this, configuration)

        JAVA_HOME_CHECK : {

            // in case JAVA_HOME is not set, but java is in the path we should not fail
            // in order to be backward compatible. Before introducing that check here
            // is worked also in case JAVA_HOME was not set, but java was in the path.
            // toolValidate works only upon JAVA_HOME and fails in case it is not set.

            def rc = sh script: 'which java' , returnStatus: true
            if(script.JAVA_HOME || (!script.JAVA_HOME && rc != 0)) {
                def java = new Tool('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
                ToolVerifier.verifyToolVersion(java, this, configuration)
            } else {
                echo 'Tool validation (java) skipped. JAVA_HOME not set, but java executable in path.'
            }
        }

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

        def mtarFileName = "${id}.mtar"
        def mtaJar = ToolUtils.getToolExecutable(mta, this, configuration)
        def buildTarget = configuration.buildTarget

        sh  """#!/bin/bash
            export PATH=./node_modules/.bin:${PATH}
            java -jar ${mtaJar} --mtar ${mtarFileName} --build-target=${buildTarget} build
            """

        def mtarFilePath = "${pwd()}/${mtarFileName}"
        script?.commonPipelineEnvironment?.setMtarFilePath(mtarFilePath)

        return mtarFilePath
    }
}

