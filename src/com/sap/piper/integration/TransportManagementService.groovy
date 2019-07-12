package com.sap.piper.integration
import com.sap.piper.JsonUtils

import com.cloudbees.groovy.cps.NonCPS

class TransportManagementService implements Serializable {

    final Script script
    final Map config

    def jsonUtils = new JsonUtils()

    TransportManagementService(Script script, Map config) {
        this.script = script
        this.config = config
    }

    def authentication(String uaaUrl, String oauthClientId, String oauthClientSecret){
        echo("OAuth Token retrieval started.")

        if(config.verbose){
            echo ("UAA-URL: '${uaaUrl}', ClientId: '${oauthClientId}''")
        }

        def encodedUsernameColonPassword = "${oauthClientId}:${oauthClientSecret}".bytes.encodeBase64().toString()
        def urlEncodedFormData = "grant_type=password&"+"username=${urlEncodeAndReplaceSpace(oauthClientId)}&"+"password=${urlEncodeAndReplaceSpace(oauthClientSecret)}"

        def parameters = [
            url          : "${uaaUrl}/oauth/token/?grant_type=client_credentials&response_type=token",
            httpMode     : "POST",
            requestBody  : urlEncodedFormData,
            customHeaders: [
                [
                    maskValue: false,
                    name     : 'Content-Type',
                    value    : 'application/x-www-form-urlencoded'
                ],
                [
                    maskValue: true,
                    name     : 'authorization',
                    value    : "Basic ${encodedUsernameColonPassword}"
                ]
            ]
        ]

        def response = sendApiRequest(parameters)
        echo ("OAuth Token retrieved successfully.")

        return jsonUtils.jsonStringToGroovyObject(response).access_token

    }



    def uploadFile(String url, String token, String file, String namedUser){

        echo("Fileupload started.")

        if(config.verbose){
            echo("URL: '${url}', File: '${file}'")
        }

        def httpResponse = script.sh returnStdout: true,
            script: """#!/bin/sh -e
                        curl -H 'Authorization: Bearer ${token}' -F 'file=@${file}' -F 'namedUser=${namedUser}' -o responseFileUpload.txt --write-out '%{http_code}' --fail '${url}/v2/files/upload'
                    """

        if(httpResponse.toInteger()  < 200 || httpResponse.toInteger() >= 300){
            script.error "[TransportManagementService] Fileupload failed. HTTP-Status: '${httpResponse}'"
        }

        def responseContent = script.readFile("responseFileUpload.txt")

        if(config.verbose){
            echo("${responseContent}")
        }

        echo("Fileupload successful.")

        return jsonUtils.jsonStringToGroovyObject(responseContent)

    }


    def uploadFileToNode(String url, String token, String nodeName, int fileId, String description, String namedUser){

        echo("Nodeupload started.")

        if(config.verbose){
            echo("URL: '${url}', Nodename: '${nodeName}', FileId: '${fileId}''")
        }

        def bodyMap = [nodeName: nodeName, contentType: 'MTA', description: description, storageType: 'FILE', namedUser: namedUser, entries: [[uri: fileId]]]

        def parameters = [
            url          : "${url}/v2/nodes/upload",
            httpMode     : "POST",
            contentType  : 'APPLICATION_JSON',
            requestBody  : jsonUtils.groovyObjectToPrettyJsonString(bodyMap),
            customHeaders: [
                [
                    maskValue: true,
                    name     : 'authorization',
                    value    : "Bearer ${token}"
                ]
            ]
        ]

        def response = sendApiRequest(parameters)
        echo("Nodeupload successful.")

        return jsonUtils.jsonStringToGroovyObject(response)

    }

    private sendApiRequest(parameters) {
        def defaultParameters = [
            acceptType            : 'APPLICATION_JSON',
            quiet                 : !config.verbose,
            consoleLogResponseBody: !config.verbose,
            ignoreSslErrors       : true,
            validResponseCodes    : "100:399"
        ]

        def response = script.httpRequest(defaultParameters + parameters)

        if (config.verbose){
            echo("Received response " + "'${response.content}' with status ${response.status}.")
        }

        return response.content
    }

    private echo(message){
        script.echo "[${getClass().getSimpleName()}] ${message}"
    }

    private static String urlEncodeAndReplaceSpace(String data) {
        return URLEncoder.encode(data, "UTF-8").replace('%20', '+')
    }
}
