import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = 'pipelineStashFilesBeforeBuild'
@Field Set STEP_CONFIG_KEYS = ['runOpaTests', 'stashIncludes', 'stashExcludes']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, stepNameDoc: 'stashFiles') {

        def utils = parameters.juStabUtils
        if (utils == null) {
            utils = new Utils()
        }

        def script = checkScript(this, parameters)
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        //additional includes via passing e.g. stashIncludes: [opa5: '**/*.include']
        //additional excludes via passing e.g. stashExcludes: [opa5: '**/*.exclude']

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME,
                               stepParam1: parameters?.script == null], config)

        if (config.runOpaTests){
            utils.stash('opa5', config.stashIncludes?.get('opa5')?config.stashIncludes.opa5:'**/*.*', config.stashExcludes?.get('opa5')?config.stashExcludes.opa5:'')
        }

        //store build descriptor files depending on technology, e.g. pom.xml, package.json
        utils.stash(
            'buildDescriptor',
            config.stashIncludes.buildDescriptor,
            config.stashExcludes.buildDescriptor
        )
        //store deployment descriptor files depending on technology, e.g. *.mtaext.yml
        utils.stashWithMessage(
            'deployDescriptor',
            '[${STEP_NAME}] no deployment descriptor files provided: ',
            config.stashIncludes.deployDescriptor,
            config.stashExcludes.deployDescriptor
        )
        //store git metadata for SourceClear agent
        sh "mkdir -p gitmetadata"
        sh "cp -rf .git/* gitmetadata"
        sh "chmod -R u+w gitmetadata"
        utils.stashWithMessage(
            'git',
            '[${STEP_NAME}] no git repo files detected: ',
            config.stashIncludes.git,
            config.stashExcludes.git
        )
        //store nsp & retire exclusion file for future use
        utils.stashWithMessage(
            'opensourceConfiguration',
            '[${STEP_NAME}] no opensourceConfiguration files provided: ',
            config.stashIncludes.get('opensourceConfiguration'),
            config.stashExcludes.get('opensourceConfiguration')
        )
        //store pipeline configuration including additional groovy test scripts for future use
        utils.stashWithMessage(
            'pipelineConfigAndTests',
            '[${STEP_NAME}] no pipeline configuration and test files found: ',
            config.stashIncludes.pipelineConfigAndTests,
            config.stashExcludes.pipelineConfigAndTests
        )
        utils.stashWithMessage(
            'securityDescriptor',
            '[${STEP_NAME}] no security descriptor found: ',
            config.stashIncludes.securityDescriptor,
            config.stashExcludes.securityDescriptor
        )
        //store files required for tests, e.g. Gauge, SUT, ...
        utils.stashWithMessage(
            'tests',
            '[${STEP_NAME}] no files for tests provided: ',
            config.stashIncludes.tests,
            config.stashExcludes.tests
        )
    }
}
