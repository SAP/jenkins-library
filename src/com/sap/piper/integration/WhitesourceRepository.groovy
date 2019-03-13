package com.sap.piper.integration

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.JsonUtils

class WhitesourceRepository implements Serializable {

    final Script script
    final Map config

    WhitesourceRepository(Script script, Map config) {
        this.script = script
        this.config = config

        if(!config.serviceUrl)
            script.error "Parameter 'serviceUrl' must be provided as part of the configuration."
    }

    List fetchVulnerabilities(whitesourceProjectsMetaInformation) {
        def fetchedVulnerabilities = []
        if (config.projectNames) {
            for (int i = 0; i < whitesourceProjectsMetaInformation.size(); i++) {
                def metaInfo = whitesourceProjectsMetaInformation[i]

                def requestBody = [
                    requestType : "getProjectAlertsByType",
                    alertType : "SECURITY_VULNERABILITY",
                    projectToken: metaInfo.token
                ]

                def response = fetchWhitesourceResource(requestBody)
                fetchedVulnerabilities.addAll(response.alerts)
            }
        } else {
            def requestBody = [
                requestType : "getProductAlertsByType",
                alertType : "SECURITY_VULNERABILITY",
                productToken: config.productToken,
            ]

            def response = fetchWhitesourceResource(requestBody)
            fetchedVulnerabilities.addAll(response.alerts)
        }

        sortVulnerabilitiesByScore(fetchedVulnerabilities)

        return fetchedVulnerabilities
    }

    protected def fetchWhitesourceResource(Map requestBody) {
        final def response = httpWhitesource(requestBody)
        def parsedResponse = new JsonUtils().parseJsonSerializable(response.content)

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
            def cvss3score1 = o1.vulnerability.cvss3_score != 0 ? o1.vulnerability.cvss3_score : o1.vulnerability.score
            def cvss3score2 = o2.vulnerability.cvss3_score != 0 ? o2.vulnerability.cvss3_score : o2.vulnerability.score

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
        if(config.projectNames){
            def requestBody = [
                requestType: "getProductProjectVitals",
                productToken: config.productToken
            ]
            def response = fetchWhitesourceResource(requestBody)

            if(response?.projectVitals) {
                projectsMetaInfo.addAll(findProjectsMeta(response.projectVitals))
            } else {
                script.error "[WhiteSource] Could not fetch any projects for product '${config.productName}' from backend, response was ${response}"
            }
        }
        return projectsMetaInfo
    }

    List findProjectsMeta(projectVitals) {
        def matchedProjects = []
        for (int i = 0; i < config.projectNames?.size(); i++) {
            def requestedProjectName = config.projectNames[i].trim()
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
                script.error "[WhiteSource] Could not fetch/find requested project '${requestedProjectName}' for product '${config.productName}'"
            }
        }

        return matchedProjects
    }

    void fetchReportForProduct(reportName) {
        def requestContent = [
            requestType: "getProductRiskReport",
            productToken: config.productToken
        ]

        fetchFileFromWhiteSource(reportName, requestContent)
    }

    def fetchProductLicenseAlerts() {
        def requestContent = [
            requestType: "getProductAlertsByType",
            alertType: "REJECTED_BY_POLICY_RESOURCE",
            productToken: config.productToken
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
    protected def httpWhitesource(requestBody) {
        handleAdditionalRequestParameters(requestBody)
        def serializedBody = new JsonUtils().getPrettyJsonString(requestBody)
        def params = [
            url        : config.serviceUrl,
            httpMode   : 'POST',
            acceptType : 'APPLICATION_JSON',
            contentType: 'APPLICATION_JSON',
            requestBody: serializedBody,
            quiet      : !config.verbose,
            timeout    : config.timeout
        ]

        if (script.env.HTTP_PROXY)
            params["httpProxy"] = script.env.HTTP_PROXY

        if(config.verbose)
            script.echo "Sending http request with parameters ${params}"

        def response = script.httpRequest(params)

        if(config.verbose)
            script.echo "Received response ${reponse}"

        return response
    }

    @NonCPS
    protected void fetchFileFromWhiteSource(String fileName, Map params) {
        handleAdditionalRequestParameters(params)
        def serializedContent = new JsonUtils().jsonToString(params)

        if(config.verbose)
            script.echo "Sending curl request with parameters ${params}"

        script.sh "${config.verbose ? '' : '#!/bin/sh -e\n'}curl -o ${fileName} -X POST ${config.serviceUrl} -H 'Content-Type: application/json' -d \'${serializedContent}\'"
    }

    @NonCPS
    protected void handleAdditionalRequestParameters(params) {
        if(config.userKey)
            params["userKey"] = config.userKey
    }
}
