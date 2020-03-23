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
        echo("OAuth Token retrieval started.")

        if (config.verbose) {
            echo("UAA-URL: '${uaaUrl}', ClientId: '${oauthClientId}''")
        }

        def encodedUsernameColonPassword = "${oauthClientId}:${oauthClientSecret}".bytes.encodeBase64().toString()
        def urlEncodedFormData = "grant_type=password&" +
            "username=${urlEncodeAndReplaceSpace(oauthClientId)}&" +
            "password=${urlEncodeAndReplaceSpace(oauthClientSecret)}"

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

        def proxy = config.proxy ? config.proxy : script.env.HTTP_PROXY

        if (proxy){
            parameters["httpProxy"] = proxy
        }

        def response = sendApiRequest(parameters)
        echo("OAuth Token retrieved successfully.")

        return jsonUtils.jsonStringToGroovyObject(response).access_token

    }


    def uploadFile(String url, String token, String file, String namedUser) {

        echo("File upload started.")

        if (config.verbose) {
            echo("URL: '${url}', File: '${file}'")
        }

        def proxy = config.proxy ? config.proxy : script.env.HTTP_PROXY

        script.sh """#!/bin/sh -e
                curl ${proxy ? '--proxy ' + proxy + ' ' : ''} -H 'Authorization: Bearer ${token}' -F 'file=@${file}' -F 'namedUser=${namedUser}' -o responseFileUpload.txt  --fail '${url}/v2/files/upload'
            """

        def responseContent = script.readFile("responseFileUpload.txt")

        if (config.verbose) {
            echo("${responseContent}")
        }

        echo("File upload successful.")

        return jsonUtils.jsonStringToGroovyObject(responseContent)

    }


    def uploadFileToNode(String url, String token, String nodeName, int fileId, String description, String namedUser) {

        echo("Node upload started.")

        if (config.verbose) {
            echo("URL: '${url}', NodeName: '${nodeName}', FileId: '${fileId}''")
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

        def proxy = config.proxy ? config.proxy : script.env.HTTP_PROXY

        if (proxy){
            parameters["httpProxy"] = proxy
        }

        def response = sendApiRequest(parameters)
        echo("Node upload successful.")

        return jsonUtils.jsonStringToGroovyObject(response)

    }

    private sendApiRequest(parameters) {
        def defaultParameters = [
            acceptType            : 'APPLICATION_JSON',
            quiet                 : !config.verbose,
            consoleLogResponseBody: false, // must be false, otherwise this reveals the api-token in the auth-request
            ignoreSslErrors       : true,
            validResponseCodes    : "100:399"
        ]

        def response = script.httpRequest(defaultParameters + parameters)

        if (config.verbose) {
            echo("Received response '${response.content}' with status ${response.status}.")
        }

        return response.content
    }

    private echo(message) {
        script.echo "[${getClass().getSimpleName()}] ${message}"
    }

    private static String urlEncodeAndReplaceSpace(String data) {
        return URLEncoder.encode(data, "UTF-8").replace('%20', '+')
    }

}
