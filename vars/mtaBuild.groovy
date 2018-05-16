import com.sap.piper.ConfigurationMerger
import com.sap.piper.MtaUtils
import com.sap.piper.tools.JavaArchiveDescriptor
import com.sap.piper.tools.ToolDescriptor


def call(Map parameters = [:]) {

    def stepName = 'mtaBuild'

    Set parameterKeys = [
        'applicationName',
        'buildTarget',
        'extension',
        'mtaJarLocation'
    ]

    Set stepConfigurationKeys = [
        'applicationName',
        'buildTarget',
        'extension',
        'mtaJarLocation'
    ]

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        prepareDefaultValues script: script

        final Map configuration = ConfigurationMerger.merge(
                                      script, stepName,
                                      parameters, parameterKeys,
                                      stepConfigurationKeys)

        def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
        java.verify(this, configuration)

        def mta = new JavaArchiveDescriptor('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v', java)
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

        def mtarFileName = "${id}.mtar"
        def mtaJar = mta.getToolExecutable(this, configuration)
        def buildTarget = configuration.buildTarget

        def mtaCall = "${mtaJar} --mtar ${mtarFileName} --build-target=${buildTarget}"

        if (configuration.extension) mtaCall += " --extension=$configuration.extension"
        mtaCall += ' build'

        sh  """#!/bin/bash
            export PATH=./node_modules/.bin:${PATH}
            $mtaCall
            """

        def mtarFilePath = "${pwd()}/${mtarFileName}"
        script?.commonPipelineEnvironment?.setMtarFilePath(mtarFilePath)

        return mtarFilePath
    }
}

