package com.sap.piper.integration

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.JsonUtils

class WhitesourceOrgAdminRepository implements Serializable {

    final Script script
    final internalWhitesource
    final Map config

    WhitesourceOrgAdminRepository(Script script, Map config) {
        this.script = script
        this.config = config
        if(!this.config.whitesource.serviceUrl && !this.config.whitesourceAccessor)
            script.error "Parameter 'whitesource.serviceUrl' must be provided as part of the configuration."
        if(this.config.whitesourceAccessor instanceof String) {
            def clazz = this.class.classLoader.loadClass(this.config.whitesourceAccessor)
            this.internalWhitesource = clazz?.newInstance(this.script, this.config)
        }
    }

    def fetchProductMetaInfo() {
        def requestBody = [
            requestType: "getOrganizationProductVitals",
            orgToken: config.whitesource.orgToken
        ]
        def parsedResponse = issueHttpRequest(requestBody)

        findProductMeta(parsedResponse)
    }

    def findProductMeta(parsedResponse) {
        def foundMetaProduct = null
        for (product in parsedResponse.productVitals) {
            if (product.name == config.whitesource.productName) {
                foundMetaProduct = product
                break
            }
        }

        return foundMetaProduct
    }

    def createProduct() {
        def requestBody = [
            requestType: "createProduct",
            orgToken: config.whitesource.orgToken,
            productName: config.whitesource.productName
        ]
        def parsedResponse = issueHttpRequest(requestBody)
        def metaInfo = parsedResponse

        def groups = []
        def users = []
        config.whitesource.emailAddressesOfInitialProductAdmins.each {
            email -> users.add(["email": email])
        }

        requestBody = [
            "requestType" : "setProductAssignments",
            "productToken" : metaInfo.productToken,
            "productMembership" : ["userAssignments":[], "groupAssignments":groups],
            "productAdmins" : ["userAssignments":users],
            "alertsEmailReceivers" : ["userAssignments":[]]
        ]
        issueHttpRequest(requestBody)

        return metaInfo
    }

    def issueHttpRequest(requestBody) {
        def response = internalWhitesource ? internalWhitesource.httpWhitesource(requestBody) : httpWhitesource(requestBody)
        def parsedResponse = new JsonUtils().parseJsonSerializable(response.content)
        if(parsedResponse?.errorCode){
            script.error "[WhiteSource] Request failed with error message '${parsedResponse.errorMessage}' (${parsedResponse.errorCode})."
        }
        return parsedResponse
    }

    @NonCPS
    protected def httpWhitesource(requestBody) {
        requestBody["userKey"] = config.whitesource.orgAdminUserKey
        def serializedBody = new JsonUtils().jsonToString(requestBody)
        def params = [
            url        : config.whitesource.serviceUrl,
            httpMode   : 'POST',
            acceptType : 'APPLICATION_JSON',
            contentType: 'APPLICATION_JSON',
            requestBody: serializedBody,
            quiet      : !config.whitesource.verbose,
            timeout    : config.whitesource.timeout
        ]

        if (script.env.HTTP_PROXY)
            params["httpProxy"] = script.env.HTTP_PROXY

        if (config.whitesource.verbose)
            script.echo "Sending http request with parameters ${params}"

        def response = script.httpRequest(params)

        if (config.whitesource.verbose)
            script.echo "Received response ${response}"

        return response
    }
}
