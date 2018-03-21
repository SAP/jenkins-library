import com.sap.piper.ConfigurationMerger
import com.sap.piper.MtaUtils

import groovy.transform.Field

@Field def DEFAULT_MTA_JAR_NAME = 'mta.jar'

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

        MTA_JAR_FILE_VALIDATE: {
            // same order like inside getMtaJar,
            def mtaJarLocation = configuration?.mtaJarLocation ?: env?.MTA_JAR_LOCATION
            def returnCodeLsMtaJar = sh script: "ls ${DEFAULT_MTA_JAR_NAME}", returnStatus:true
            if(mtaJarLocation || ( !mtaJarLocation && returnCodeLsMtaJar != 0)) {
                // toolValidate commented since it is does not work in
                // conjunction with jenkins slaves.
                // TODO: switch on again when the issue is resolved.
                // toolValidate tool: 'mta', home: mtaJarLocation
                echo 'toolValidate (mta) is disabled.'
            } else {
                echo "mta toolset (${DEFAULT_MTA_JAR_NAME}) has been found in current working directory. Using this version without further tool validation."
            }
        }

        JAVA_HOME_CHECK : {

            // in case JAVA_HOME is not set, but java is in the path we should not fail
            // in order to be backward compatible. Before introducing that check here
            // is worked also in case JAVA_HOME was not set, but java was in the path.
            // toolValidate works only upon JAVA_HOME and fails in case it is not set.

            def rc = sh script: 'which java' , returnStatus: true
            if(script.JAVA_HOME || (!script.JAVA_HOME && rc != 0)) {
                // toolValidate commented since it is does not work in
                // conjunction with jenkins slaves.
                // TODO: switch on again when the issue is resolved.
                echo 'toolValidate (mta) is disabled.'
                // toolValidate tool: 'java', home: script.JAVA_HOME
                echo 'toolValidate (java) is disabled.'
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

        def mtaJar = getMtaJar(stepName, configuration)
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

private getMtaJar(stepName, configuration) {
    def mtaJarLocation = DEFAULT_MTA_JAR_NAME //default, maybe it is in current working directory

    if(configuration?.mtaJarLocation){
        mtaJarLocation = "${configuration.mtaJarLocation}/${DEFAULT_MTA_JAR_NAME}"
        echo "[$stepName] MTA JAR \"${mtaJarLocation}\" retrieved from configuration."
        return mtaJarLocation
    }

    if(env?.MTA_JAR_LOCATION){
        mtaJarLocation = "${env.MTA_JAR_LOCATION}/${DEFAULT_MTA_JAR_NAME}"
        echo "[$stepName] MTA JAR \"${mtaJarLocation}\" retrieved from environment."
        return mtaJarLocation
    }

    echo "[$stepName] Using MTA JAR from current working directory."
    return mtaJarLocation
}