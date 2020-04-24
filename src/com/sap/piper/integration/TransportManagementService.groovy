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
        if (response.status != 200) {
            prepareAndThrowException(response, "OAuth Token retrieval failed (HTTP status code '${response.status}').")
        }

        echo("OAuth Token retrieved successfully.")
        return jsonUtils.jsonStringToGroovyObject(response.content).access_token
    }


    def uploadFile(String url, String token, String file, String namedUser) {

        echo("File upload started.")

        if (config.verbose) {
            echo("URL: '${url}', File: '${file}'")
        }

        def proxy = config.proxy ? config.proxy : script.env.HTTP_PROXY

        def responseFileUpload = 'responseFileUpload.txt'

        def responseContent

        def responseCode = script.sh returnStdout: true,
                                      script:"""|#!/bin/sh -e
                                                | curl ${proxy ? '--proxy ' + proxy + ' ' : ''} \\
                                                |      --write-out '%{response_code}' \\
                                                |      -H 'Authorization: Bearer ${token}' \\
                                                |      -F 'file=@${file}' \\
                                                |      -F 'namedUser=${namedUser}' \\
                                                |      --output ${responseFileUpload} \\
                                                |      '${url}/v2/files/upload'""".stripMargin()


        def responseBody = 'n/a'

        boolean gotResponse = script.fileExists(responseFileUpload)

        if(gotResponse) {
            responseBody = script.readFile(responseFileUpload)
            if(config.verbose) {
                echo("Response body: ${responseBody}")
            }
        }

        def HTTP_CREATED = '201'

        if (responseCode != HTTP_CREATED) {
            def message = "Unexpected response code received from file upload (${responseCode}). ${HTTP_CREATED} expected."
            echo "${message} Response body: ${responseBody}"
            script.error message
        }

        echo("File upload successful.")

        if (! gotResponse) {
            script.error "Cannot provide upload file response."
        }
        return jsonUtils.jsonStringToGroovyObject(responseBody)
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
        if (response.status != 200) {
            prepareAndThrowException(response, "Node upload failed (HTTP status code '${response.status}').")
        }

        def successMessage = "Node upload successful."
        if (config.verbose) {
            successMessage += " Response content '${response.content}'."
        }
        echo(successMessage)
        return jsonUtils.jsonStringToGroovyObject(response.content)
    }

    private sendApiRequest(parameters) {
        def defaultParameters = [
            acceptType            : 'APPLICATION_JSON',
            quiet                 : !config.verbose,
            consoleLogResponseBody: false, // must be false, otherwise this reveals the api-token in the auth-request
            ignoreSslErrors       : true,
            validResponseCodes    : "100:599"
        ]

        return script.httpRequest(defaultParameters + parameters)
    }

    private prepareAndThrowException(response, errorMessage) {
        if (response.status >= 300) {
            errorMessage += " Response content '${response.content}'."
        }
        script.error "[${getClass().getSimpleName()}] ${errorMessage}"
    }

    private echo(message) {
        script.echo "[${getClass().getSimpleName()}] ${message}"
    }

    private static String urlEncodeAndReplaceSpace(String data) {
        return URLEncoder.encode(data, "UTF-8").replace('%20', '+')
    }

}
