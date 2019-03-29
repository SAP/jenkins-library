package com.sap.piper.integration

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.JsonUtils

class WhitesourceRepository implements Serializable {

    final Script script
    final Map config

    WhitesourceRepository(Script script, Map config) {
        this.script = script
        this.config = config

        if(!config?.whitesource?.serviceUrl)
            script.error "Parameter 'whitesource.serviceUrl' must be provided as part of the configuration."
    }

    List fetchVulnerabilities(whitesourceProjectsMetaInformation) {
        def fetchedVulnerabilities = []
        if (config.whitesource.projectNames) {
            for (int i = 0; i < whitesourceProjectsMetaInformation.size(); i++) {
                fetchSecurityAlertsPerItem(whitesourceProjectsMetaInformation[i].token, "getProjectAlertsByType", fetchedVulnerabilities)
            }
        } else {
            fetchSecurityAlertsPerItem(config.whitesource.productToken, "getProductAlertsByType", fetchedVulnerabilities)
        }

        sortVulnerabilitiesByScore(fetchedVulnerabilities)

        return fetchedVulnerabilities
    }

    private fetchSecurityAlertsPerItem(token, type, List<Object> fetchedVulnerabilities) {
        def requestBody = [
            requestType : type,
            alertType   : "SECURITY_VULNERABILITY",
            projectToken: token
        ]

        def response = fetchWhitesourceResource(requestBody)
        fetchedVulnerabilities.addAll(response.alerts)
    }

    protected def fetchWhitesourceResource(Map requestBody) {
        final def response = httpWhitesource(requestBody)
        def parsedResponse = new JsonUtils().jsonStringToGroovyObject(response.content)

        if(parsedResponse?.errorCode){
            script.error "[WhiteSource] Request failed with error message '${parsedResponse.errorMessage}' (${parsedResponse.errorCode})."
        }

        return parsedResponse
    }

    @NonCPS
    void sortLibrariesAlphabeticallyGAV(List libraries) {
        script.echo "found a total of ${libraries.size()} dependencies (direct and indirect)"
        libraries.sort { o1, o2 ->
            String groupID1 = o1.groupId
            String groupID2 = o2.groupId
            def comparisionResult = groupID1 <=> groupID2;

            if (comparisionResult != 0) {
                comparisionResult
            } else {
                String artifactID1 = o1.artifactId
                String artifactID2 = o2.artifactId

                artifactID1 <=> artifactID2
            }
        }
    }

    @NonCPS
    void sortVulnerabilitiesByScore(List vulnerabilities) {
        script.echo "${vulnerabilities.size() > 0 ? 'WARNING: ' : ''}found a total of ${vulnerabilities.size()} vulnerabilities"
        vulnerabilities.sort { o1, o2 ->
            def cvss3score1 = o1.vulnerability.cvss3_score == 0 ? o1.vulnerability.score : o1.vulnerability.cvss3_score
            def cvss3score2 = o2.vulnerability.cvss3_score == 0 ? o2.vulnerability.score : o2.vulnerability.cvss3_score

            def comparisionResult = cvss3score1 <=> cvss3score2

            if (comparisionResult != 0) {
                -comparisionResult
            } else {
                def score1 = o1.vulnerability.score
                def score2 = o2.vulnerability.score

                -(score1 <=> score2)
            }
        }
    }

    List fetchProjectsMetaInfo() {
        def projectsMetaInfo = []
        if(config.whitesource.projectNames){
            def requestBody = [
                requestType: "getProductProjectVitals",
                productToken: config.whitesource.productToken
            ]
            def response = fetchWhitesourceResource(requestBody)

            if(response?.projectVitals) {
                projectsMetaInfo.addAll(findProjectsMeta(response.projectVitals))
            } else {
                script.error "[WhiteSource] Could not fetch any projects for product '${config.whitesource.productName}' from backend, response was ${response}"
            }
        }
        return projectsMetaInfo
    }

    List findProjectsMeta(projectVitals) {
        def matchedProjects = []
        for (int i = 0; i < config.whitesource.projectNames?.size(); i++) {
            def requestedProjectName = config.whitesource.projectNames[i].trim()
            def matchedProjectInfo = null

            for (int j = 0; j < projectVitals.size(); j++) {
                def projectResponse = projectVitals[j]
                if (projectResponse.name == requestedProjectName) {
                    matchedProjectInfo = projectResponse
                    break
                }
            }

            if (matchedProjectInfo != null) {
                matchedProjects.add(matchedProjectInfo)
            } else {
                script.error "[WhiteSource] Could not fetch/find requested project '${requestedProjectName}' for product '${config.whitesource.productName}'"
            }
        }

        return matchedProjects
    }

    void fetchReportForProduct(reportName) {
        def headers = [[name: 'Cache-Control', value: 'no-cache, no-store, must-revalidate'], [name: 'Pragma', value: 'no-cache']]
        def requestContent = [
            requestType: "getProductRiskReport",
            productToken: config.whitesource.productToken
        ]

        //fetchFileFromWhiteSource(reportName, requestContent)
        httpWhitesource(requestContent, 'APPLICATION_OCTETSTREAM', headers, reportName)
    }

    def fetchProductLicenseAlerts() {
        def requestContent = [
            requestType: "getProductAlertsByType",
            alertType: "REJECTED_BY_POLICY_RESOURCE",
            productToken: config.whitesource.productToken
        ]
        def parsedResponse = fetchWhitesourceResource(requestContent)

        return parsedResponse
    }

    def fetchProjectLicenseAlerts(String projectToken) {
        def requestContent = [
            requestType: "getProjectAlertsByType",
            alertType: "REJECTED_BY_POLICY_RESOURCE",
            projectToken: projectToken
        ]
        def parsedResponse = fetchWhitesourceResource(requestContent)

        return parsedResponse
    }

    @NonCPS
    protected def httpWhitesource(requestBody, acceptType = 'APPLICATION_JSON', customHeaders = null, outputFile = null) {
        handleAdditionalRequestParameters(requestBody)
        def serializedBody = new JsonUtils().groovyObjectToPrettyJsonString(requestBody)
        def params = [
            url        : config.whitesource.serviceUrl,
            httpMode   : 'POST',
            acceptType : acceptType,
            contentType: 'APPLICATION_JSON',
            requestBody: serializedBody,
            quiet      : !config.verbose,
            timeout    : config.whitesource.timeout
        ]

        if(customHeaders) params["customHeaders"] = customHeaders

        if (outputFile) params["outputFile"] = outputFile

        if (script.env.HTTP_PROXY) params["httpProxy"] = script.env.HTTP_PROXY

        if(config.verbose)
            script.echo "Sending http request with parameters ${params}"

        def response = script.httpRequest(params)

        if(config.verbose)
            script.echo "Received response ${response}"

        return response
    }

    @NonCPS
    protected void handleAdditionalRequestParameters(params) {
        if(config.whitesource.userKey)
            params["userKey"] = config.whitesource.userKey
    }
}
