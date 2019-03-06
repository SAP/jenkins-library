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
    /**
     * Jenkins credentials ID referring to the organization admin's token.
     */
    'orgAdminUserTokenCredentialsId',
    /**
     * WhiteSource token identifying your organization.
     */
    'orgToken',
    /**
     * Name of the WhiteSource product to be created and used for results aggregation.
     */
    'productName',
    /**
     * Version of the WhiteSource product to be created and used for results aggregation, usually determined automatically.
     */
    'productVersion',
    /**
     * Token of the WhiteSource product to be created and used for results aggregation, usually determined automatically.
     */
    'productToken',
    /**
     * List of WhiteSource projects to be included in the assessment part of the step, usually determined automatically.
     */
    'projectNames',
    /**
     * Type of development stack used to implement the solution.
     * @possibleValues `maven`, `mta`, `npm`, `pip`, `sbt`
     */
    'scanType',
    /**
     * URL to the WhiteSource server API used for communication, defaults to `https://saas.whitesourcesoftware.com/api`.
     */
    'serviceUrl',
    /**
     * Jenkins credentials ID referring to the product admin's token.
     */
    'userTokenCredentialsId',
    /**
     * Whether verbose output should be produced.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS + [
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
     * URL used for downloading the Java Runtime Environment (JRE) required to run the WhiteSource Unified Agent.
     */
    'jreDownloadUrl',
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
    whitesource   : [
        orgAdminUserTokenCredentialsId   : 'orgAdminUserTokenCredentialsId',
        orgToken                         : 'orgToken',
        productName                      : 'productName',
        productVersion                   : 'productVersion',
        productToken                     : 'productToken',
        projectNames                     : 'projectNames',
        scanType                         : 'scanType',
        serviceUrl                       : 'serviceUrl',
        userTokenCredentialsId           : 'userTokenCredentialsId'
    ],
    whitesourceUserTokenCredentialsId    : 'userTokenCredentialsId',
    whitesourceProductName               : 'productName',
    whitesourceProjectNames              : 'projectNames',
    whitesourceProductToken              : 'productToken'
]

/**
 * With this step [WhiteSource](https://www.whitesource.com) security and license compliance scans can be executed and assessed.
 *
 * WhiteSource is a Software as a Service offering based on a so called unified scanning agent that locally determines the dependency
 * tree of a node.js, Java, Python, Ruby, or Scala based solution and sends it to the WhiteSource server for a policy based license compliance
 * check and additional Free and Open Source Software Publicly Known Vulnerabilities assessment.
 *
 * !!! note "Docker Images"
 *     The underlying Docker images are public and specific to the solution's programming language(s) and may therefore be exchanged
 *     to fit and suite the relevant scenario. The default Python environment used is i.e. Python 3 based.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def descriptorUtils = parameters.descriptorUtilsStub ?: new DescriptorUtils()
        def statusCode = 1

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin([
                style : libraryResource('piper-os.css')
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('scanType').mixin('buildDescriptorFile')
            .dependingOn('scanType').mixin('configFilePath')
            .dependingOn('scanType').mixin('dockerImage')
            .dependingOn('scanType').mixin('dockerWorkspace')
            .dependingOn('scanType').mixin('stashContent')
            .withMandatoryProperty('orgToken')
            .withMandatoryProperty('userTokenCredentialsId')
            .withMandatoryProperty('productName')
            .use()

        config.cvssSeverityLimit = config.cvssSeverityLimit == null ? -1 : Integer.valueOf(config.cvssSeverityLimit)
        config.stashContent = utils.unstashAll(config.stashContent)
        config.projectNames = (config.projectNames instanceof List) ? config.projectNames : config.projectNames?.tokenize(',')
        parameters.projectNames = config.projectNames

        script.commonPipelineEnvironment.setInfluxStepData('whitesource', false)

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scanType',
            stepParam1: config.scanType
        ], config)

        echo "Parameters: scanType: ${config.scanType}"

        def whitesourceRepository = parameters.whitesourceRepositoryStub ?: new WhitesourceRepository(this, config)
        def whitesourceOrgAdminRepository = parameters.whitesourceOrgAdminRepositoryStub ?: new WhitesourceOrgAdminRepository(this, config)

        statusCode = triggerWhitesourceScanWithUserKey(script, config, utils, descriptorUtils, parameters, whitesourceRepository, whitesourceOrgAdminRepository)

        checkStatus(statusCode, config)

        script.commonPipelineEnvironment.setInfluxStepData('whitesource', true)
    }
}

private def triggerWhitesourceScanWithUserKey(script, config, utils, descriptorUtils, parameters, repository, orgAdminRepository) {
    withCredentials ([string(
        credentialsId: config.userTokenCredentialsId,
        variable: 'userKey'
    )]) {
        config.userKey = userKey
        def statusCode = 1
        echo "Triggering Whitesource scan on product '${config.productName}' with token '${config.productToken}' using credentials with ID '${config.userTokenCredentialsId}'"
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
                if (config.parallelLimit > 0 && config.parallelLimit < scanJobs.keySet().size()) {
                    // block wise
                    def scanJobsAll = scanJobs
                    scanJobs = [failFast: false]
                    for (int i = 1; i <= scanJobsAll.keySet().size(); i++) {
                        def index = i - 1
                        def key = scanJobsAll.keySet()[index]
                        scanJobs[key] = scanJobsAll[key]
                        if (i % config.parallelLimit == 0 || i == scanJobsAll.keySet().size()) {
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
                def gav
                switch (config.scanType) {
                    case 'npm':
                        gav = descriptorUtils.getNpmGAV(config.buildDescriptorFile)
                        config.projectName = gav.group + "." + gav.artifact
                        config.productVersion = gav.version
                        break
                    case 'sbt':
                        gav = descriptorUtils.getSbtGAV(config.buildDescriptorFile)
                        config.projectName = gav.group + "." + gav.artifact
                        config.productVersion = gav.version
                        break
                    case 'pip':
                        gav = descriptorUtils.getPipGAV(config.buildDescriptorFile)
                        config.projectName = gav.artifact
                        config.productVersion = gav.version
                        break
                    default:
                        gav = descriptorUtils.getMavenGAV(config.buildDescriptorFile)
                        config.projectName = gav.group + "." + gav.artifact
                        config.productVersion = gav.version
                        break
                }
                config.projectNames.add("${config.projectName} - ${config.productVersion}".toString())
                WhitesourceConfigurationHelper.extendUAConfigurationFile(script, utils, config, path)
                dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                    if (config.agentDownloadUrl) {
                        def agentDownloadUrl = new GStringTemplateEngine().createTemplate(config.agentDownloadUrl).make([config: config]).toString()
                        //if agentDownloadUrl empty, rely on dockerImage to contain unifiedAgent correctly set up and available
                        sh "curl ${script.env.HTTP_PROXY ? '--proxy ' + script.env.HTTP_PROXY + ' ' : ''}--location --output ${config.agentFileName} ${agentDownloadUrl}".toString()
                    }

                    def javaCmd = 'java'
                    if (config.jreDownloadUrl) {
                        //if jreDownloadUrl empty, rely on dockerImage to contain java correctly set up and available on the path
                        sh "curl ${script.env.HTTP_PROXY ? '--proxy ' + script.env.HTTP_PROXY + ' ' : ''}--location --output jvm.tar.gz ${config.jreDownloadUrl} && tar --strip-components=1 -xzf jvm.tar.gz".toString()
                        javaCmd = './bin/java'
                    }

                    def options = ["-jar ${config.agentFileName} -c \'${config.configFilePath}\'"]
                    if (config.orgToken) options.push("-apiKey '${config.orgToken}'")
                    if (config.userKey) options.push("-userKey '${config.userKey}'")
                    if (config.productName) options.push("-product '${config.productName}'")

                    statusCode = sh(script: "${javaCmd} ${options.join(' ')} ${config.agentParameters}", returnStatus: true)

                    if (config.agentDownloadUrl) {
                        sh "rm -f ${config.agentFileName}"
                    }

                    if (config.jreDownloadUrl) {
                        sh "rm -rf ./bin ./conf ./legal ./lib ./man"
                    }

                    // archive whitesource result files
                    archiveArtifacts artifacts: "whitesource/*.*", allowEmptyArchive: true
                }
                break
        }

        if (config.reporting) {
            analyseWhitesourceResults(config, repository, orgAdminRepository)
        }

        return statusCode
    }
}

void analyseWhitesourceResults(Map config, WhitesourceRepository repository, WhitesourceOrgAdminRepository orgAdminRepository) {
    if (!config.productToken) {
        def metaInfo = orgAdminRepository.fetchProductMetaInfo()
        def key = "token"
        if(!metaInfo && config.createProductFromPipeline) {
            metaInfo = orgAdminRepository.createProduct()
            key = "productToken"
        } else if(!metaInfo) {
            error "[WhiteSource] Could not fetch/find requested product '${config.productName}' and automatic creation has been disabled"
        }
        echo "Meta Info: ${metaInfo}"
        config.productToken = metaInfo[key]
    }

    def pdfName = "whitesource-riskReport.pdf"
    repository.fetchReportForProduct(pdfName)
    archiveArtifacts artifacts: pdfName
    echo "A summary of the Whitesource findings was stored as artifact under the name ${pdfName}"

    if(config.licensingVulnerabilities) {
        def violationCount = fetchViolationCount(config, repository)
        checkViolationStatus(violationCount)
    }

    if (config.securityVulnerabilities)
        config.severeVulnerabilities = checkSecurityViolations(config, repository)
}

int fetchViolationCount(Map config, WhitesourceRepository repository) {
    int violationCount = 0
    if (config.projectNames) {
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
    def whitesourceProjectsMetaInformation = repository.fetchProjectsMetaInfo()
    def whitesourceVulnerabilities = repository.fetchVulnerabilities(whitesourceProjectsMetaInformation)
    def severeVulnerabilities = 0
    whitesourceVulnerabilities.each {
        item ->
            if ((item.vulnerability.score >= config.cvssSeverityLimit || item.vulnerability.cvss3_score >= config.cvssSeverityLimit) && config.cvssSeverityLimit >= 0)
                severeVulnerabilities++
    }

    writeFile(file: "${config.vulnerabilityReportFileName}.json", text: new JsonUtils().getPrettyJsonString(whitesourceVulnerabilities))
    writeFile(file: "${config.vulnerabilityReportFileName}.html", text: getReportHtml(config, whitesourceVulnerabilities, severeVulnerabilities))
    archiveArtifacts(artifacts: "${config.vulnerabilityReportFileName}.*")

    if (whitesourceVulnerabilities.size() - severeVulnerabilities > 0)
        echo "[${STEP_NAME}] WARNING: ${whitesourceVulnerabilities.size() - severeVulnerabilities} Open Source Software Security vulnerabilities with CVSS score below ${config.cvssSeverityLimit} detected."
    if (whitesourceVulnerabilities.size() == 0)
        echo "[${STEP_NAME}] No Open Source Software Security vulnerabilities detected."

    return severeVulnerabilities
}

// ExitCodes: https://whitesource.atlassian.net/wiki/spaces/WD/pages/34209870/NPM+Plugin#NPMPlugin-ExitCode
void checkStatus(int statusCode, config) {
    def errorMessage = ""
    if(config.securityVulnerabilities && config.severeVulnerabilities > 0)
        errorMessage += "${config.severeVulnerabilities} Open Source Software Security vulnerabilities with CVSS score greater or equal ${config.cvssSeverityLimit} detected. - "
    if (config.licensingVulnerabilities)
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
    def now = new Date().format('MMM dd, yyyy - HH:mm:ss')
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
            reportTitle                 : config.vulnerabilityReportTitle,
            style                       : config.style,
            totalSevereVulnerabilities  : numSevereVulns,
            totalVulnerabilities        : vulnerabilityList.size(),
            vulnerabilityTable          : vulnerabilityTable,
            whitesourceProductName      : config.productName,
            whitesourceProjectNames     : config.projectNames
        ]).toString()
}
