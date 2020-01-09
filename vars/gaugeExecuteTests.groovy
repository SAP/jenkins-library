import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.analytics.InfluxData
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = [
    /**
     * Defines the build tool to be used for the test execution.
     * @possibleValues `maven`, `npm`, `bundler`
     */
    'buildTool',
    /** @see dockerExecute*/
    'dockerEnvVars',
    /** @see dockerExecute*/
    'dockerImage',
    /** @see dockerExecute*/
    'dockerName',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute*/
    'dockerWorkspace',
    /**
     * Defines the behavior in case tests fail. When this is set to `true` test results cannot be recorded using the `publishTestResults` step afterwards.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /** Defines the command for installing Gauge. In case the `dockerImage` already contains Gauge it can be set to empty: ``.*/
    'installCommand',
    /** Defines the Gauge language runner to be used.*/
    'languageRunner',
    /** Defines the command which is used for executing Gauge.*/
    'runCommand',
    /** Defines if specific stashes should be considered for the tests.*/
    'stashContent',
    /** Allows to set specific options for the Gauge execution. Details can be found for example [in the Gauge Maven plugin documentation](https://github.com/getgauge/gauge-maven-plugin#executing-specs)*/
    'testOptions',
    /** Defines the repository containing the tests, in case the test implementation is stored in a different repository than the code itself.*/
    'testRepository',
    /** Defines the branch containing the tests, in case the test implementation is stored in a different repository and a different branch than master.*/
    'gitBranch',
    /**
     * Defines the credentials for the repository containing the tests, in case the test implementation is stored in a different and protected repository than the code itself.
     * For protected repositories the `testRepository` needs to contain the ssh git url.
     */
    'gitSshKeyCredentialsId',
    /** It is passed as environment variable `TARGET_SERVER_URL` to the test execution. Tests running against the system should read the host information from this environment variable in order to be infrastructure agnostic.*/
    'testServerUrl'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this step Gauge ([getgauge.io](http:getgauge.io)) acceptance tests are executed.
 * Using Gauge it will be possible to have a three-tier test layout:
 *
 * * Acceptance Criteria
 * * Test implemenation layer
 * * Application driver layer
 *
 * This layout is propagated by Jez Humble and Dave Farley in their book "Continuous Delivery" as a way to create maintainable acceptance test suites (see "Continuous Delivery", p. 190ff).
 *
 * Using Gauge it is possible to write test specifications in [Markdown syntax](http://daringfireball.net/projects/markdown/syntax) and therefore allow e.g. product owners to write the relevant acceptance test specifications. At the same time it allows the developer to implement the steps described in the specification in her development environment.
 *
 * You can use the [sample projects](https://github.com/getgauge/gauge-mvn-archetypes) of Gauge.
 *
 * !!! note "Make sure to run against a Selenium Hub configuration"
 *     In the test example of _gauge-archetype-selenium_ please make sure to allow it to run against a Selenium hub:
 *
 *     Please extend DriverFactory.java for example in following way:
 *
 *     ``` java
 *     String hubUrl = System.getenv("HUB_URL");
 *     //when running on a Docker deamon (and not using Kubernetes plugin), Docker images will be linked
 *     //in this case hubUrl will be http://selenium:4444/wd/hub due to the linking of the containers
 *     hubUrl = (hubUrl == null) ? "http://localhost:4444/wd/hub" : hubUrl;
 *     Capabilities chromeCapabilities = DesiredCapabilities.chrome();
 *     System.out.println("Running on Selenium Hub: " + hubUrl);
 *     return new RemoteWebDriver(new URL(hubUrl), chromeCapabilities);
 *     ```
*/
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters)  ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        InfluxData.addField('step_data', 'gauge', false)

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildTool').mixin('dockerImage')
            .dependingOn('buildTool').mixin('dockerName')
            .dependingOn('buildTool').mixin('dockerOptions')
            .dependingOn('buildTool').mixin('dockerEnvVars')
            .dependingOn('buildTool').mixin('dockerWorkspace')
            .dependingOn('buildTool').mixin('languageRunner')
            .dependingOn('buildTool').mixin('runCommand')
            .dependingOn('buildTool').mixin('testOptions')
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'buildTool',
            stepParam1: config.buildTool,
            stepParamKey2: 'dockerName',
            stepParam2: config.dockerName
        ], config)

        if(!config.dockerEnvVars.TARGET_SERVER_URL && config.testServerUrl)
            config.dockerEnvVars.TARGET_SERVER_URL = config.testServerUrl

        if (config.testRepository) {
            // handle separate test repository
            config.stashContent = [GitUtils.handleTestRepository(this, config)]
        } else {
            config.stashContent = utils.unstashAll(config.stashContent)
        }

        seleniumExecuteTests (
            script: script,
            buildTool: config.buildTool,
            dockerImage: config.dockerImage,
            dockerName: config.dockerName,
            dockerEnvVars: config.dockerEnvVars,
            dockerOptions: config.dockerOptions,
            dockerWorkspace: config.dockerWorkspace,
            stashContent: config.stashContent
        ) {
            String gaugeScript = ''
            if (config.installCommand) {
                gaugeScript = '''export HOME=${HOME:-$(pwd)}
                    if [ "$HOME" = "/" ]; then export HOME=$(pwd); fi
                    export PATH=$HOME/bin/gauge:$PATH
                    mkdir -p $HOME/bin/gauge
                    ''' + config.installCommand + '''
                    gauge telemetry off
                    gauge install ''' + config.languageRunner + '''
                    gauge install html-report
                    gauge install xml-report
                    '''
            }
            gaugeScript += config.runCommand

            try {
                sh "${gaugeScript} ${config.testOptions}"
                InfluxData.addField('step_data', 'gauge', true)
            } catch (err) {
                echo "[${STEP_NAME}] One or more tests failed"
                script.currentBuild.result = 'UNSTABLE'
                if (config.failOnError) throw err
            }
        }
    }
}
