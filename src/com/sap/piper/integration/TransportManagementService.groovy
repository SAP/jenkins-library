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

        return jsonUtils.jsonStringToGroovyObject(getResponseBody(responseFileUpload, responseCode, '201', 'File upload'))
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

    def uploadMtaExtDescriptorToNode(String url, String token, Long nodeId, String file, String mtaVersion, String description, String namedUser) {

        echo("MTA Extension Descriptor upload started.")

        if (config.verbose) {
            echo("URL: '${url}', NodeId: '${nodeId}', File: '${file}', MtaVersion: '${mtaVersion}'")
        }

        def proxy = config.proxy ? config.proxy : script.env.HTTP_PROXY

        def responseExtDescriptorUpload = 'responseExtDescriptorUpload.txt'

        def responseCode = script.sh returnStdout: true,
                                     script: """|#!/bin/sh -e
                                                | curl ${proxy ? '--proxy ' + proxy + ' ' : ''} \\
                                                |      --write-out '%{response_code}' \\
                                                |      -H 'Authorization: Bearer ${token}' \\
                                                |      -H 'tms-named-user: ${namedUser}' \\
                                                |      -F 'file=@${file}' \\
                                                |      -F 'mtaVersion=${mtaVersion}' \\
                                                |      -F 'description=${description}' \\
                                                |      --output ${responseExtDescriptorUpload} \\
                                                |      '${url}/v2/nodes/${nodeId}/mtaExtDescriptors'""".stripMargin()

        return jsonUtils.jsonStringToGroovyObject(getResponseBody(responseExtDescriptorUpload, responseCode, '201', 'MTA Extension Descriptor upload'))
    }

    def getNodes(String url, String token) {

        if (config.verbose) {
            echo("Get nodes started. URL: '${url}'")
        }

        def parameters = [
            url          : "${url}/v2/nodes",
            httpMode     : "GET",
            contentType  : 'APPLICATION_JSON',
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
            prepareAndThrowException(response, "Get nodes failed (HTTP status code '${response.status}').")
        }

        if (config.verbose) {
            echo("Get nodes successful. Response content '${response.content}'.")
        }

        return jsonUtils.jsonStringToGroovyObject(response.content)
    }

    def updateMtaExtDescriptor(String url, String token, Long nodeId, Long idOfMtaExtDescriptor, String file, String mtaVersion, String description, String namedUser) {

        echo("MTA Extension Descriptor update started.")

        if (config.verbose) {
        echo("URL: '${url}', NodeId: '${nodeId}', IdOfMtaDescriptor: '${idOfMtaExtDescriptor}', File: '${file}', MtaVersion: '${mtaVersion}'")
        }

        def proxy = config.proxy ? config.proxy : script.env.HTTP_PROXY

        def responseExtDescriptorUpdate = 'responseExtDescriptorUpdate.txt'

        def responseCode = script.sh returnStdout: true,
                                     script: """|#!/bin/sh -e
                                                | curl ${proxy ? '--proxy ' + proxy + ' ' : ''} \\
                                                |      --write-out '%{response_code}' \\
                                                |      -H 'Authorization: Bearer ${token}' \\
                                                |      -H 'tms-named-user: ${namedUser}' \\
                                                |      -F 'file=@${file}' \\
                                                |      -F 'mtaVersion=${mtaVersion}' \\
                                                |      -F 'description=${description}' \\
                                                |      --output ${responseExtDescriptorUpdate} \\
                                                |      -X PUT \\
                                                |      '${url}/v2/nodes/${nodeId}/mtaExtDescriptors/${idOfMtaExtDescriptor}'""".stripMargin()

        return jsonUtils.jsonStringToGroovyObject(getResponseBody(responseExtDescriptorUpdate, responseCode, '200', 'MTA Extension Descriptor update'))
    }

    def getMtaExtDescriptor(String url, String token, Long nodeId, String mtaId, String mtaVersion) {
        echo("Get MTA Extension Descriptor started.")

        if (config.verbose) {
            echo("URL: '${url}', NodeId: '${nodeId}', MtaId: '${mtaId}', MtaVersion: '${mtaVersion}'")
        }

        def parameters = [
            url          : "${url}/v2/nodes/${nodeId}/mtaExtDescriptors?mtaId=${mtaId}&mtaVersion=${mtaVersion}",
            httpMode     : "GET",
            contentType  : 'APPLICATION_JSON',
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
            prepareAndThrowException(response, "Get MTA Extension Descriptor failed (HTTP status code '${response.status}').")
        }

        if (config.verbose) {
            echo("Response content '${response.content}'.")
        }

        // because the API is called with params, the return is always a map with either a empty list or a list containing one element
        Map mtaExtDescriptor = [:]
        Map responseContent = jsonUtils.jsonStringToGroovyObject(response.content)
        def mtaExtDescriptors = responseContent.get("mtaExtDescriptors")
        if(mtaExtDescriptors && mtaExtDescriptors.size() > 0) {
            mtaExtDescriptor = mtaExtDescriptors.get(0)
        }

        echo("Get MTA Extension Descriptor successful.")
        return mtaExtDescriptor
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


    private String getResponseBody(String responseFileName, String actualStatus, String expectStatus, String action) {
        def responseBody = 'n/a'

        boolean gotResponse = script.fileExists(responseFileName)
        if(gotResponse) {
            responseBody = script.readFile(responseFileName)
            if(config.verbose) {
                echo("Response body: ${responseBody}")
            }
        }

        if (actualStatus != expectStatus) {
            def message = "Unexpected response code received from ${action} (${actualStatus}). ${expectStatus} expected."
            echo "${message} Response body: ${responseBody}"
            script.error message
        }

        echo("${action} successful.")

        if (! gotResponse) {
            script.error "Cannot provide response for ${action}."
        }
        return responseBody
    }
}
