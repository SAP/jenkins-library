package com.sap.piper.integration

import com.sap.piper.JsonUtils

class TransportManagementService implements Serializable {

    final Script script
    final Map config

    def jsonUtils = new JsonUtils()

    TransportManagementService(Script script, Map config) {

        this.script = script
        this.config = config

    }

    def authentication(String uaaUrl, String oauthClientId, String oauthClientSecret) {

        script.echo "[TransportManagementService] OAuth Token retrieval started."

        if (config.verbose) {

            script.echo "[TransportManagementService] UAA-URL: '${uaaUrl}', ClientId: '${oauthClientId}'"

        }

        def httpResponse = script.sh returnStdout: true,
            script: """#!/bin/sh -e
                        curl -XPOST -u '${oauthClientId}':'${oauthClientSecret}' -o responseAuth.txt --write-out '%{http_code}' '${uaaUrl}/oauth/token/?grant_type=client_credentials&response_type=token'
                    """

        def response = script.readFile("responseAuth.txt")

        if (httpResponse.toInteger() < 200 || httpResponse.toInteger() >= 300) {

            script.error "[TransportManagementService] Retrieval of OAuth-Token failed. HTTP-Status: '${httpResponse}' \n [ERROR] Response: '${response}'"

        }

        if (config.verbose) {

            script.echo response

        }

        def oAuthToken = jsonUtils.jsonStringToGroovyObject(response).access_token

        script.echo "[TransportManagementService] OAuth Token retrieved successfully."

        return oAuthToken

    }


    def uploadFileToTMS(String url, String token, String file, String namedUser) {

        script.echo "[TransportManagementService] Fileupload started."

        if (config.verbose) {

            script.echo "[TransportManagementService] URL: '${url}', File: '${file}'"

        }

        def httpResponse = script.sh returnStdout: true,
            script: """#!/bin/sh -e
                        curl -XPOST -H 'Authorization: Bearer ${token}' -F 'file=@${file}' -F 'namedUser=${namedUser}' -o responseFileUpload.txt --write-out '%{http_code}' '${url}/v2/files/upload'
                    """

        def response = script.readFile("responseFileUpload.txt")

        if (httpResponse.toInteger() < 200 || httpResponse.toInteger() >= 300) {

            script.error "[TransportManagementService] Fileupload failed. HTTP-Status: '${httpResponse}' \n [ERROR] Response: '${response}'"

        }

        if (config.verbose) {

            script.echo response

        }

        def fileUploadDetails = jsonUtils.jsonStringToGroovyObject(response)

        script.echo "[TransportManagementService] Fileupload successful."

        return fileUploadDetails

    }


    def uploadFileToNode(String url, String token, String nodeName, int fileId, String description, String namedUser) {

        script.echo "[TransportManagementService] Nodeupload started."

        if (config.verbose) {

            script.echo "[TransportManagementService] URL: '${url}', Nodename: '${nodeName}', FileId: '${fileId}'"

        }

        def bodyMap = [nodeName: nodeName, contentType: 'MTA', description: description, storageType: 'FILE', namedUser: namedUser, entries: [[uri: fileId]]]

        def body = jsonUtils.groovyObjectToPrettyJsonString(bodyMap)

        def httpResponse = script.sh returnStdout: true,
            script: """#!/bin/sh -e
                        curl -XPOST -H 'Authorization: Bearer ${token}' -H 'Content-Type: application/json' -d '${body}' -o responseNodeUpload.txt --write-out '%{http_code}' '${url}/v2/nodes/upload'
                    """

        def response = script.readFile("responseNodeUpload.txt")

        if (httpResponse.toInteger() < 200 || httpResponse.toInteger() >= 300) {

            script.error "[TransportManagementService] Nodeupload failed. HTTP-Status: ${httpResponse} \n [ERROR] Response: ${response}"

        }

        if (config.verbose) {

            script.echo response

        }

        def nodeUploadDetails = jsonUtils.jsonStringToGroovyObject(response)

        script.echo "[TransportManagementService] Nodeupload successful."

        return nodeUploadDetails

    }

}
