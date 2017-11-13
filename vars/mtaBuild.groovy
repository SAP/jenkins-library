import com.sap.piper.Utils

/**
 * mtaBuild
 * Builds Fiori app with Multitarget Archiver
 * Prerequisite: InitializeNpm needs to be called beforehand
 *
 */
def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'mtaBuild', stepParameters: parameters) {

        def utils = new Utils()
        def buildTarget = utils.getMandatoryParameter(parameters, 'buildTarget', null)
        def script = parameters.script
        if (script == null){
            script = [commonPipelineEnvironment: commonPipelineEnvironment]
        }

        def mtaYaml = readYaml file: "${pwd()}/mta.yaml"

        //[Q]: Why not yaml.dump()? [A]: This reformats the whole file.
        sh "sed -ie \"s/\\\${timestamp}/`date +%Y%m%d%H%M%S`/g\" \"${pwd()}/mta.yaml\""

        def id = mtaYaml.ID
        if (!id) {
            error "Property 'ID' not found in mta.yaml file at: '${pwd()}'"
        }

        def mtarFileName = "${id}.mtar"

        def mtaJar = getMtaJar(parameters)

        sh  """#!/bin/bash
            export PATH=./node_modules/.bin:${PATH}
            java -jar ${mtaJar} --mtar ${mtarFileName} --build-target=${buildTarget} build
            """

        def mtarFilePath = "${pwd()}/${mtarFileName}"
        script.commonPipelineEnvironment.setMtarFilePath(mtarFilePath)

        return mtarFilePath
    }
}

private getMtaJar(parameters) {
    def mtaJarLocation = 'mta.jar' //default, maybe it is in current working directory

    if(parameters?.mtaJarLocation){
        mtaJarLocation = "${parameters.mtaJarLocation}/mta.jar"
        echo "[mtaBuild] MTA JAR \"${mtaJarLocation}\" retrieved from parameters."
        return mtaJarLocation
    }

    if(env?.MTA_JAR_LOCATION){
        mtaJarLocation = "${env.MTA_JAR_LOCATION}/mta.jar"
        echo "[mtaBuild] MTA JAR \"${mtaJarLocation}\" retrieved from environment."
        return mtaJarLocation
    }

    echo "[mtaBuild] Using MTA JAR from current working directory."
    return mtaJarLocation
}
