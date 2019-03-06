package com.sap.piper.integration

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.JsonUtils

class WhitesourceOrgAdminRepository implements Serializable {

    final Script script
    final internalWhitesource
    final Map config

    def orgAdminUserKey

    WhitesourceOrgAdminRepository(Script script, Map config) {
        this.script = script
        this.config = config
        if(!this.config.serviceUrl && !this.config.whitesourceAccessor)
            script.error "Parameter 'serviceUrl' must be provided as part of the configuration."
        if(this.config.whitesourceAccessor instanceof String) {
            def clazz = this.class.classLoader.loadClass(this.config.whitesourceAccessor)
            this.internalWhitesource = clazz?.newInstance(this.script, this.config)
        }
    }

    def fetchProductMetaInfo() {
        def requestBody = [
            requestType: "getOrganizationProductVitals",
            orgToken: config.orgToken
        ]
        def response = internalWhitesource ? internalWhitesource.httpWhitesource(requestBody) : httpWhitesource(requestBody)
        def parsedResponse = new JsonUtils().parseJsonSerializable(response.content)

        findProductMeta(parsedResponse)
    }

    def findProductMeta(parsedResponse) {
        def foundMetaProduct = null
        for (product in parsedResponse.productVitals) {
            if (product.name == config.productName) {
                foundMetaProduct = product
                break
            }
        }

        return foundMetaProduct
    }

    def createProduct() {
        def requestBody = [
            requestType: "createProduct",
            orgToken: config.orgToken,
            productName: config.productName
        ]
        def response = issueHttpRequest(requestBody)
        def parsedResponse = new JsonUtils().parseJsonSerializable(response.content)
        def metaInfo = parsedResponse

        def groups = []
        def users = []
        config.emailAddressesOfInitialProductAdmins.each {
            email -> users.add(["email": config.emailOfInitialProductAdmin])
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
        internalWhitesource ? internalWhitesource.httpWhitesource(requestBody) : httpWhitesource(requestBody)
    }

    @NonCPS
    protected def httpWhitesource(requestBody) {
        script.withCredentials ([script.string(
            credentialsId: config.orgAdminUserTokenCredentialsId,
            variable: 'orgAdminUserKey'
        )]) {
            requestBody["userKey"] = orgAdminUserKey
            def serializedBody = new JsonUtils().jsonToString(requestBody)
            def params = [
                url        : config.serviceUrl,
                httpMode   : 'POST',
                acceptType : 'APPLICATION_JSON',
                contentType: 'APPLICATION_JSON',
                requestBody: serializedBody,
                quiet      : !config.verbose,
                timeout    : config.timeout
            ]

            if (script.env.HTTP_PROXY && !config.serviceUrl.matches('http(s)*://.*\\.sap\\.corp.*'))
                params["httpProxy"] = script.env.HTTP_PROXY

            if (config.verbose)
                script.echo "Sending http request with parameters ${params}"

            def response = script.httpRequest(params)

            if (config.verbose)
                script.echo "Received response ${response}"

            return response
        }
    }
}
