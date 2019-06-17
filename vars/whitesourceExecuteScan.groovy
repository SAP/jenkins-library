import com.sap.piper.DescriptorUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JsonUtils
import com.sap.piper.Utils
import com.sap.piper.integration.WhitesourceOrgAdminRepository
import com.sap.piper.integration.WhitesourceRepository
import com.sap.piper.ConfigurationHelper
import com.sap.piper.WhitesourceConfigurationHelper
import com.sap.piper.mta.MtaMultiplexer
import groovy.text.GStringTemplateEngine
import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    'whitesource',
    /**
     * Jenkins credentials ID referring to the organization admin's token.
     * @parentConfigKey whitesource
     */
    'orgAdminUserTokenCredentialsId',
    /**
     * WhiteSource token identifying your organization.
     * @parentConfigKey whitesource
     */
    'orgToken',
    /**
     * Name of the WhiteSource product to be created and used for results aggregation.
     * @parentConfigKey whitesource
     */
    'productName',
    /**
     * Version of the WhiteSource product to be created and used for results aggregation, usually determined automatically.
     * @parentConfigKey whitesource
     */
    'productVersion',
    /**
     * Token of the WhiteSource product to be created and used for results aggregation, usually determined automatically.
     * @parentConfigKey whitesource
     */
    'productToken',
    /**
     * List of WhiteSource projects to be included in the assessment part of the step, usually determined automatically.
     * @parentConfigKey whitesource
     */
    'projectNames',
    /**
     * URL used for downloading the Java Runtime Environment (JRE) required to run the WhiteSource Unified Agent.
     * @parentConfigKey whitesource
     */
    'jreDownloadUrl',
    /**
     * URL to the WhiteSource server API used for communication, defaults to `https://saas.whitesourcesoftware.com/api`.
     * @parentConfigKey whitesource
     */
    'serviceUrl',
    /**
     * Jenkins credentials ID referring to the product admin's token.
     * @parentConfigKey whitesource
     */
    'userTokenCredentialsId',
    /**
     * Type of development stack used to implement the solution.
     * @possibleValues `golang`, `maven`, `mta`, `npm`, `pip`, `sbt`
     */
    'scanType',
    /**
     * Whether verbose output should be produced.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS + [
    /**
     * Install command that can be used to populate the default docker image for some scenarios.
     */
    'installCommand',
    /**
     * URL used to download the latest version of the WhiteSource Unified Agent.
     */
    'agentDownloadUrl',
    /**
     * Locally used name for the Unified Agent jar file after download.
     */
    'agentFileName',
    /**
     * Additional parameters passed to the Unified Agent command line.
     */
    'agentParameters',
    /**
     * List of build descriptors and therefore modules to exclude from the scan and assessment activities.
     */
    'buildDescriptorExcludeList',
    /**
     * Explicit path to the build descriptor file.
     */
    'buildDescriptorFile',
    /**
     * Explicit path to the WhiteSource Unified Agent configuration file.
     */
    'configFilePath',
    /**
     * Whether to create the related WhiteSource product on the fly based on the supplied pipeline configuration.
     */
    'createProductFromPipeline',
    /**
     * The list of email addresses to assign as product admins for newly created WhiteSource products.
     */
    'emailAddressesOfInitialProductAdmins',
    /**
     * Docker image to be used for scanning.
     */
    'dockerImage',
    /**
     * Docker workspace to be used for scanning.
     */
    'dockerWorkspace',
    /**
     * Whether license compliance is considered and reported as part of the assessment.
     * @possibleValues `true`, `false`
     */
    'licensingVulnerabilities',
    /**
     * Limit of parallel jobs being run at once in case of `scanType: 'mta'` based scenarios, defaults to `15`.
     */
    'parallelLimit',
    /**
     * Whether assessment is being done at all, defaults to `true`.
     * @possibleValues `true`, `false`
     */
    'reporting',
    /**
     * Whether security compliance is considered and reported as part of the assessment.
     * @possibleValues `true`, `false`
     */
    'securityVulnerabilities',
    /**
     * Limit of tollerable CVSS v3 score upon assessment and in consequence fails the build, defaults to  `-1`.
     * @possibleValues `-1` to switch failing off, any `positive integer between 0 and 10` to fail on issues with the specified limit or above
     */
    'cvssSeverityLimit',
    /**
     * List of stashes to be unstashed into the workspace before performing the scan.
     */
    'stashContent',
    /**
     * Timeout in seconds until a HTTP call is forcefully terminated.
     */
    'timeout',
    /**
     * Name of the file the vulnerability report is written to.
     */
    'vulnerabilityReportFileName',
    /**
     * Title of vulnerability report written during the assessment phase.
     */
    'vulnerabilityReportTitle'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@Field Map CONFIG_KEY_COMPATIBILITY = [
    productName                        : 'whitesourceProductName',
    productToken                       : 'whitesourceProductToken',
    projectNames                       : 'whitesourceProjectNames',
    userTokenCredentialsId             : 'whitesourceUserTokenCredentialsId',
    serviceUrl                         : 'whitesourceServiceUrl',
    agentDownloadUrl                   : 'fileAgentDownloadUrl',
    agentParameters                    : 'fileAgentParameters',
    whitesource                        : [
        orgAdminUserTokenCredentialsId          : 'orgAdminUserTokenCredentialsId',
        orgToken                                : 'orgToken',
        productName                             : 'productName',
        productToken                            : 'productToken',
        projectNames                            : 'projectNames',
        productVersion                          : 'productVersion',
        serviceUrl                              : 'serviceUrl',
        configFilePath                          : 'configFilePath',
        userTokenCredentialsId                  : 'userTokenCredentialsId',
        agentDownloadUrl                        : 'agentDownloadUrl',
        agentFileName                           : 'agentFileName',
        agentParameters                         : 'agentParameters',
        buildDescriptorExcludeList              : 'buildDescriptorExcludeList',
        buildDescriptorFile                     : 'buildDescriptorFile',
        createProductFromPipeline               : 'createProductFromPipeline',
        emailAddressesOfInitialProductAdmins    : 'emailAddressesOfInitialProductAdmins',
        jreDownloadUrl                          : 'jreDownloadUrl',
        licensingVulnerabilities                : 'licensingVulnerabilities',
        parallelLimit                           : 'parallelLimit',
        reporting                               : 'reporting',
        securityVulnerabilities                 : 'securityVulnerabilities',
        cvssSeverityLimit                       : 'cvssSeverityLimit',
        timeout                                 : 'timeout',
        vulnerabilityReportFileName             : 'vulnerabilityReportFileName',
        vulnerabilityReportTitle                : 'vulnerabilityReportTitle',
        installCommand                          : 'installCommand'
    ]
]

/**
 * BETA
 *
 * With this step [WhiteSource](https://www.whitesourcesoftware.com) security and license compliance scans can be executed and assessed.
 *
 * WhiteSource is a Software as a Service offering based on a so called unified agent that locally determines the dependency
 * tree of a node.js, Java, Python, Ruby, or Scala based solution and sends it to the WhiteSource server for a policy based license compliance
 * check and additional Free and Open Source Software Publicly Known Vulnerabilities detection.
 *
 * !!! note "Docker Images"
 *     The underlying Docker images are public and specific to the solution's programming language(s) and therefore may have to be exchanged
 *     to fit to and support the relevant scenario. The default Python environment used is i.e. Python 3 based.
 *
 * !!! warn "Restrictions"
 *     Currently the step does contain hardened scan configurations for `scanType` `'pip'` and `'go'`. Other environments are still being elaborated,
 *     so please thoroughly check your results and do not take them for granted by default.
 *     Also not all environments have been thoroughly tested already therefore you might need to tweak around with the default containers used or
 *     create your own ones to adequately support your scenario. To do so please modify `dockerImage` and `dockerWorkspace` parameters.
 *     The step expects an environment containing the programming language related compiler/interpreter as well as the related build tool. For a list
 *     of the supported build tools per environment please refer to the [WhiteSource Unified Agent Documentation](https://whitesource.atlassian.net/wiki/spaces/WD/pages/33718339/Unified+Agent).
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def descriptorUtils = parameters.descriptorUtilsStub ?: new DescriptorUtils()
        def statusCode = 1

        //initialize CPE for passing whiteSourceProjects
        if(script.commonPipelineEnvironment.getValue('whitesourceProjectNames') == null) {
            script.commonPipelineEnvironment.setValue('whitesourceProjectNames', [])
        }

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults(CONFIG_KEY_COMPATIBILITY)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin([
                style : libraryResource('piper-os.css')
            ])
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .dependingOn('scanType').mixin('buildDescriptorFile')
            .dependingOn('scanType').mixin('dockerImage')
            .dependingOn('scanType').mixin('dockerWorkspace')
            .dependingOn('scanType').mixin('stashContent')
            .dependingOn('scanType').mixin('whitesource/configFilePath')
            .dependingOn('scanType').mixin('whitesource/installCommand')
            .withMandatoryProperty('whitesource/serviceUrl')
            .withMandatoryProperty('whitesource/orgToken')
            .withMandatoryProperty('whitesource/userTokenCredentialsId')
            .withMandatoryProperty('whitesource/productName')
            .use()

        config.whitesource.cvssSeverityLimit = config.whitesource.cvssSeverityLimit == null ?: Integer.valueOf(config.whitesource.cvssSeverityLimit)
        config.stashContent = utils.unstashAll(config.stashContent)
        config.whitesource['projectNames'] = (config.whitesource['projectNames'] instanceof List) ? config.whitesource['projectNames'] : config.whitesource['projectNames']?.tokenize(',')
        parameters.whitesource = parameters.whitesource ?: [:]
        parameters.whitesource['projectNames'] = config.whitesource['projectNames']

        script.commonPipelineEnvironment.setInfluxStepData('whitesource', false)

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scanType',
            stepParam1: config.scanType
        ], config)

        echo "Parameters: scanType: ${config.scanType}"

        def whitesourceRepository = parameters.whitesourceRepositoryStub ?: new WhitesourceRepository(this, config)
        def whitesourceOrgAdminRepository = parameters.whitesourceOrgAdminRepositoryStub ?: new WhitesourceOrgAdminRepository(this, config)

        if(config.whitesource.orgAdminUserTokenCredentialsId) {
            statusCode = triggerWhitesourceScanWithOrgAdminUserKey(script, config, utils, descriptorUtils, parameters, whitesourceRepository, whitesourceOrgAdminRepository)
        } else {
            statusCode = triggerWhitesourceScanWithUserKey(script, config, utils, descriptorUtils, parameters, whitesourceRepository, whitesourceOrgAdminRepository)
        }
        checkStatus(statusCode, config)

        script.commonPipelineEnvironment.setInfluxStepData('whitesource', true)
    }
}

private def triggerWhitesourceScanWithOrgAdminUserKey(script, config, utils, descriptorUtils, parameters, repository, orgAdminRepository) {
    withCredentials ([script.string(
        credentialsId: config.whitesource.orgAdminUserTokenCredentialsId,
        variable: 'orgAdminUserKey'
    )]) {
        config.whitesource.orgAdminUserKey = orgAdminUserKey
        triggerWhitesourceScanWithUserKey(script, config, utils, descriptorUtils, parameters, repository, orgAdminRepository)
    }
}

private def triggerWhitesourceScanWithUserKey(script, config, utils, descriptorUtils, parameters, repository, orgAdminRepository) {
    withCredentials ([string(
        credentialsId: config.whitesource.userTokenCredentialsId,
        variable: 'userKey'
    )]) {
        config.whitesource.userKey = userKey
        def statusCode = 1
        echo "Triggering Whitesource scan on product '${config.whitesource.productName}'${config.whitesource.productToken ? ' with token \'' + config.whitesource.productToken + '\'' : ''} using product admin credentials with ID '${config.whitesource.userTokenCredentialsId}'${config.whitesource.orgAdminUserTokenCredentialsId ? ' and organization admin credentials with ID \'' + config.whitesource.orgAdminUserTokenCredentialsId + '\'' : ''}"

        if (!config.whitesource.productToken) {
            def metaInfo = orgAdminRepository.fetchProductMetaInfo()
            def key = "token"
            if((null == metaInfo || !metaInfo[key]) && config.whitesource.createProductFromPipeline) {
                metaInfo = orgAdminRepository.createProduct()
                key = "productToken"
            } else if(null == metaInfo || !metaInfo[key]) {
                error "[WhiteSource] Could not fetch/find requested product '${config.whitesource.productName}' and automatic creation has been disabled"
            }
            echo "Meta Info: ${metaInfo}"
            config.whitesource.productToken = metaInfo[key]
        }

        switch (config.scanType) {
            case 'mta':
                def scanJobs = [:]
                def mtaParameters = [:] + parameters + [reporting: false]
                // harmonize buildDescriptorExcludeList
                config.buildDescriptorExcludeList = config.buildDescriptorExcludeList instanceof List ? config.buildDescriptorExcludeList : config.buildDescriptorExcludeList?.replaceAll(', ', ',').replaceAll(' ,', ',').tokenize(',')
                // create job for each pom.xml with scanType: 'maven'
                scanJobs.putAll(MtaMultiplexer.createJobs(
                    this, mtaParameters, config.buildDescriptorExcludeList, 'Whitesource', 'pom.xml', 'maven'
                ) { options -> return whitesourceExecuteScan(options) })
                // create job for each pom.xml with scanType: 'maven'
                scanJobs.putAll(MtaMultiplexer.createJobs(
                    this, mtaParameters, config.buildDescriptorExcludeList, 'Whitesource', 'package.json', 'npm'
                ) { options -> whitesourceExecuteScan(options) })
                // create job for each setup.py with scanType: 'pip'
                scanJobs.putAll(MtaMultiplexer.createJobs(
                    this, mtaParameters, config.buildDescriptorExcludeList, 'Whitesource', 'setup.py', 'pip'
                ) { options -> whitesourceExecuteScan(options) })
                // execute scan jobs
                if (config.whitesource.parallelLimit > 0 && config.whitesource.parallelLimit < scanJobs.keySet().size()) {
                    // block wise
                    def scanJobsAll = scanJobs
                    scanJobs = [failFast: false]
                    for (int i = 1; i <= scanJobsAll.keySet().size(); i++) {
                        def index = i - 1
                        def key = scanJobsAll.keySet()[index]
                        scanJobs[key] = scanJobsAll[key]
                        if (i % config.whitesource.parallelLimit == 0 || i == scanJobsAll.keySet().size()) {
                            parallel scanJobs
                            scanJobs = [failFast: false]
                        }
                    }
                } else {
                    // in parallel
                    scanJobs += [failFast: false]
                    parallel scanJobs
                }
                statusCode = 0
                break
            default:
                def path = config.buildDescriptorFile.substring(0, config.buildDescriptorFile.lastIndexOf('/') + 1)
                resolveProjectIdentifiers(script, descriptorUtils, config)

                def projectName = "${config.whitesource.projectName}${config.whitesource.productVersion?' - ':''}${config.whitesource.productVersion?:''}".toString()
                if(!config.whitesource['projectNames'].contains(projectName))
                    config.whitesource['projectNames'].add(projectName)

                //share projectNames with other steps
                if(!script.commonPipelineEnvironment.getValue('whitesourceProjectNames').contains(projectName))
                    script.commonPipelineEnvironment.getValue('whitesourceProjectNames').add(projectName)

                WhitesourceConfigurationHelper.extendUAConfigurationFile(script, utils, config, path)
                dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                    if (config.whitesource.agentDownloadUrl) {
                        def agentDownloadUrl = new GStringTemplateEngine().createTemplate(config.whitesource.agentDownloadUrl).make([config: config]).toString()
                        //if agentDownloadUrl empty, rely on dockerImage to contain unifiedAgent correctly set up and available
                        sh "curl ${script.env.HTTP_PROXY ? '--proxy ' + script.env.HTTP_PROXY + ' ' : ''}--location --output ${config.whitesource.agentFileName} ${agentDownloadUrl}".toString()
                    }

                    def javaCmd = 'java'
                    if (config.whitesource.jreDownloadUrl) {
                        //if jreDownloadUrl empty, rely on dockerImage to contain java correctly set up and available on the path
                        sh "curl ${script.env.HTTP_PROXY ? '--proxy ' + script.env.HTTP_PROXY + ' ' : ''}--location --output jvm.tar.gz ${config.whitesource.jreDownloadUrl} && tar --strip-components=1 -xzf jvm.tar.gz".toString()
                        javaCmd = './bin/java'
                    }

                    if(config.whitesource.installCommand)
                        sh new GStringTemplateEngine().createTemplate(config.whitesource.installCommand).make([config: config]).toString()

                    def options = ["-jar ${config.whitesource.agentFileName} -c \'${config.whitesource.configFilePath}\'"]
                    if (config.whitesource.orgToken) options.push("-apiKey '${config.whitesource.orgToken}'")
                    if (config.whitesource.userKey) options.push("-userKey '${config.whitesource.userKey}'")
                    if (config.whitesource.productName) options.push("-product '${config.whitesource.productName}'")

                    statusCode = sh(script: "${javaCmd} ${options.join(' ')} ${config.whitesource.agentParameters}", returnStatus: true)

                    if (config.whitesource.agentDownloadUrl) {
                        sh "rm -f ${config.whitesource.agentFileName}"
                    }

                    if (config.whitesource.jreDownloadUrl) {
                        sh "rm -rf ./bin ./conf ./legal ./lib ./man"
                        sh "rm -f jvm.tar.gz"
                    }

                    // archive whitesource result files for UA
                    archiveArtifacts artifacts: "whitesource/*.*", allowEmptyArchive: true

                    // archive whitesource debug files, if available
                    archiveArtifacts artifacts: "**/ws-l*", allowEmptyArchive: true
                }
                break
        }

        if (config.reporting) {
            analyseWhitesourceResults(config, repository)
        }

        return statusCode
    }
}

private resolveProjectIdentifiers(script, descriptorUtils, config) {
    if (!config.whitesource.projectName || !config.whitesource.productVersion) {
        def gav
        switch (config.scanType) {
            case 'npm':
                gav = descriptorUtils.getNpmGAV(config.buildDescriptorFile)
                break
            case 'sbt':
                gav = descriptorUtils.getSbtGAV(config.buildDescriptorFile)
                break
            case 'pip':
                gav = descriptorUtils.getPipGAV(config.buildDescriptorFile)
                break
            case 'golang':
                gav = descriptorUtils.getGoGAV(config.buildDescriptorFile, new URI(script.commonPipelineEnvironment.getGitHttpsUrl()))
                break
            case 'dlang':
                break
            case 'maven':
                gav = descriptorUtils.getMavenGAV(config.buildDescriptorFile)
                break
        }

        if(!config.whitesource.projectName)
            config.whitesource.projectName = "${gav.group?:''}${gav.group?'.':''}${gav.artifact}"

        def versionFragments = gav.version?.tokenize('.')
        def version = versionFragments.size() > 0 ? versionFragments.head() : null
        if(version && !config.whitesource.productVersion)
            config.whitesource.productVersion = version
    }
}

void analyseWhitesourceResults(Map config, WhitesourceRepository repository) {
    def pdfName = "whitesource-riskReport.pdf"
    repository.fetchReportForProduct(pdfName)
    archiveArtifacts artifacts: pdfName
    echo "A summary of the Whitesource findings was stored as artifact under the name ${pdfName}"

    if(config.whitesource.licensingVulnerabilities) {
        def violationCount = fetchViolationCount(config, repository)
        checkViolationStatus(violationCount)
    }

    if (config.whitesource.securityVulnerabilities)
        config.whitesource.severeVulnerabilities = checkSecurityViolations(config, repository)
}

int fetchViolationCount(Map config, WhitesourceRepository repository) {
    int violationCount = 0
    if (config.whitesource?.projectNames) {
        def projectsMeta = repository.fetchProjectsMetaInfo()
        for (int i = 0; i < projectsMeta.size(); i++) {
            def project = projectsMeta[i]
            def responseAlertsProject = repository.fetchProjectLicenseAlerts(project.token)
            violationCount += responseAlertsProject.alerts.size()
        }
    } else {
        def responseAlerts = repository.fetchProductLicenseAlerts()
        violationCount += responseAlerts.alerts.size()
    }
    return violationCount
}

void checkViolationStatus(int violationCount) {
    if (violationCount == 0) {
        echo "[${STEP_NAME}] No policy violations found"
    } else {
        error "[${STEP_NAME}] Whitesource found ${violationCount} policy violations for your product"
    }
}

int checkSecurityViolations(Map config, WhitesourceRepository repository) {
    def projectsMetaInformation = repository.fetchProjectsMetaInfo()
    def vulnerabilities = repository.fetchVulnerabilities(projectsMetaInformation)
    def severeVulnerabilities = 0
    vulnerabilities.each {
        item ->
            if ((item.vulnerability.score >= config.whitesource.cvssSeverityLimit || item.vulnerability.cvss3_score >= config.whitesource.cvssSeverityLimit) && config.whitesource.cvssSeverityLimit >= 0)
                severeVulnerabilities++
    }

    writeFile(file: "${config.vulnerabilityReportFileName}.json", text: new JsonUtils().groovyObjectToPrettyJsonString(vulnerabilities))
    writeFile(file: "${config.vulnerabilityReportFileName}.html", text: getReportHtml(config, vulnerabilities, severeVulnerabilities))
    archiveArtifacts(artifacts: "${config.vulnerabilityReportFileName}.*")

    if (vulnerabilities.size() - severeVulnerabilities > 0)
        echo "[${STEP_NAME}] WARNING: ${vulnerabilities.size() - severeVulnerabilities} Open Source Software Security vulnerabilities with CVSS score below ${config.whitesource.cvssSeverityLimit} detected."
    if (vulnerabilities.size() == 0)
        echo "[${STEP_NAME}] No Open Source Software Security vulnerabilities detected."

    return severeVulnerabilities
}

// ExitCodes: https://whitesource.atlassian.net/wiki/spaces/WD/pages/34209870/NPM+Plugin#NPMPlugin-ExitCode
void checkStatus(int statusCode, config) {
    def errorMessage = ""
    if(config.whitesource.securityVulnerabilities && config.whitesource.severeVulnerabilities > 0)
        errorMessage += "${config.whitesource.severeVulnerabilities} Open Source Software Security vulnerabilities with CVSS score greater or equal ${config.whitesource.cvssSeverityLimit} detected. - "
    if (config.whitesource.licensingVulnerabilities)
        switch (statusCode) {
            case 0:
                break
            case 255:
                errorMessage += "The scan resulted in an error"
                break
            case 254:
                errorMessage += "Whitesource found one or multiple policy violations"
                break
            case 253:
                errorMessage += "The local scan client failed to execute the scan"
                break
            case 252:
                errorMessage += "There was a failure in the connection to the WhiteSource servers"
                break
            case 251:
                errorMessage += "The server failed to analyze the scan"
                break
            case 250:
                errorMessage += "Pre-step failure"
                break
            default:
                errorMessage += "Whitesource scan failed with unknown error code '${statusCode}'"
        }

    if (errorMessage)
        error "[${STEP_NAME}] " + errorMessage
}

def getReportHtml(config, vulnerabilityList, numSevereVulns) {
    def now = new Date().format('MMM dd, yyyy - HH:mm:ss z', TimeZone.getTimeZone('UTC'))
    def vulnerabilityTable = ''
    if (vulnerabilityList.size() == 0) {
        vulnerabilityTable += '''
            <tr>
                <td colspan=12> No publicly known vulnerabilities detected </td>
            </tr>'''
    } else {
        for (int i = 0; i < vulnerabilityList.size(); i++) {
            def item = vulnerabilityList[i]
            def score = item.vulnerability.cvss3_score > 0 ? item.vulnerability.cvss3_score : item.vulnerability.score
            def topFix = item.vulnerability.topFix ? "${item.vulnerability.topFix?.message}<br>${item.vulnerability.topFix?.fixResolution}<br><a href=\"${item.vulnerability.topFix?.url}\">${item.vulnerability.topFix?.url}</a>}" : ''
            vulnerabilityTable += """
            <tr>
                <td>${i + 1}</td>
                <td>${item.date}</td>
                <td><a href=\"${item.vulnerability.url}\">${item.vulnerability.name}</a></td>
                <td class=\"${score < config.cvssSeverityLimit ? 'warn' : 'notok'}\">${score}</td>
                <td>${item.vulnerability.cvss3_score > 0 ? 'v3' : 'v2'}</td>
                <td>${item.project}</td>
                <td>${item.library.filename}</td>
                <td>${item.library.groupId}</td>
                <td>${item.library.artifactId}</td>
                <td>${item.library.version}</td>
                <td>${item.vulnerability.description}</td>
                <td>${topFix}</td>
            </tr>"""
        }
    }

    return SimpleTemplateEngine.newInstance().createTemplate(libraryResource('com.sap.piper/templates/whitesourceVulnerabilities.html')).make(
        [
            now                         : now,
            reportTitle                 : config.whitesource.vulnerabilityReportTitle,
            style                       : config.style,
            cvssSeverityLimit           : config.whitesource.cvssSeverityLimit,
            totalSevereVulnerabilities  : numSevereVulns,
            totalVulnerabilities        : vulnerabilityList.size(),
            vulnerabilityTable          : vulnerabilityTable,
            whitesourceProductName      : config.whitesource.productName,
            whitesourceProjectNames     : config.whitesource.projectNames
        ]).toString()
}
