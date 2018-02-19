import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.ToolUtils


def call(Map parameters = [:]) {

    def stepName = 'mtaBuild'

    List parameterKeys = [
        'buildTarget',
        'mtaJarLocation'
    ]

    List stepConfigurationKeys = [
        'buildTarget'
    ]

    List generalConfigurationKeys = [
        'mtaJarLocation'
    ]

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        prepareDefaultValues script: script

        final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)
        final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, stepName)
        final Map generalConfiguration = ConfigurationLoader.generalConfiguration(script)
        final Map configuration = ConfigurationMerger.merge(
                                      parameters, parameterKeys,
                                      generalConfiguration, generalConfigurationKeys, [:],
                                      stepConfiguration, stepConfigurationKeys, stepDefaults)

        def mtaYaml = readYaml file: "${pwd()}/mta.yaml"

        //[Q]: Why not yaml.dump()? [A]: This reformats the whole file.
        sh "sed -ie \"s/\\\${timestamp}/`date +%Y%m%d%H%M%S`/g\" \"${pwd()}/mta.yaml\""

        def id = mtaYaml.ID
        if (!id) {
            error "Property 'ID' not found in mta.yaml file at: '${pwd()}'"
        }

        def mtarFileName = "${id}.mtar"
        def mtaJar = ToolUtils.getMtaJar(this, 'mtaBuild', configuration, env)
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

