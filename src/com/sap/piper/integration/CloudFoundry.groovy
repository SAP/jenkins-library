package com.sap.piper.integration

class CloudFoundry {
    final Script script

    CloudFoundry(Script script) {
        this.script = script
    }

    def getXsuaaCredentials(String apiEndpoint, String org, String space, String credentialsId, String appName, boolean verbose){
        def env = getAppEnvironment(apiEndpoint, org, space, credentialsId, appName, verbose)
        if (!env.system_env_json.VCAP_SERVICES.xsuaa[0]) throw new Exception("Cannot find xsuaa credentials")
        return env.system_env_json.VCAP_SERVICES.xsuaa[0].credentials
    }

    def getAppEnvironment(String apiEndpoint, String org, String space, String credentialsId, String appName, boolean verbose){
        def authEndpoint = getAuthEndPoint(apiEndpoint, verbose)
        def bearerToken = getBearerToken(authEndpoint, credentialsId, verbose)
        def appUrl = getAppRefUrl(apiEndpoint, org, space, bearerToken, appName, verbose)
        def response = script.httpRequest url: "${appUrl}/env", quiet: !verbose,
                        customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        def envJson = script.readJSON text:"${response.content}"
        return envJson
    }

    def getAuthEndPoint(String apiEndpoint, boolean verbose){
        def info = script.httpRequest url: "${apiEndpoint}/v2/info", quiet: !verbose
        def responseJson = script.readJSON text:"${info.content}"
        return responseJson.authorization_endpoint
    }

    def getBearerToken(String authorizationEndpoint, String credentialsId, boolean verbose){
        script.withCredentials([script.usernamePassword(credentialsId: credentialsId, usernameVariable: 'usercf', passwordVariable: 'passwordcf')]) {
            def token = script.httpRequest url:"${authorizationEndpoint}/oauth/token", quiet: !verbose,
                                httpMode:'POST',
                                requestBody: "username=${script.usercf}&password=${script.passwordcf}&client_id=cf&grant_type=password&response_type=token",
                                customHeaders: [[name: 'Content-Type', value: 'application/x-www-form-urlencoded'],[name: 'Authorization', value: 'Basic Y2Y6']]
            def responseJson = script.readJSON text:"${token.content}"
            return "Bearer ${responseJson.access_token.trim()}"
        }
    }

    def getAppRefUrl(String apiEndpoint, String org, String space, String bearerToken, String appName, boolean verbose){
        def orgGuid = getOrgGuid(apiEndpoint, org, bearerToken, verbose)
        def spaceGuid = getSpaceGuid(apiEndpoint, orgGuid, space, bearerToken, verbose)
        def appInfo = script.httpRequest  url:"${apiEndpoint}/v3/apps?names=${appName},${appName}_blue,${appName}_green,${appName}-blue,${appName}-green&space_guids=${spaceGuid}",  quiet: !verbose,
                            customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        def responseJson = script.readJSON text:"${appInfo.content}"
        if (!responseJson.resources[0]) throw new Exception("Cannot find Application")
        return responseJson.resources[0].links.self.href.trim()
    }

    def getOrgGuid(String apiEndpoint, String org, String bearerToken, boolean verbose){
        def organizationInfo = script.httpRequest url: "${apiEndpoint}/v3/organizations?names=${org}", quiet: !verbose,
                                    customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        def responseJson = script.readJSON text:"${organizationInfo.content}"
        if (!responseJson.resources[0]) throw new Exception("Cannot find Organization")
        return responseJson.resources[0].guid
    }

    def getSpaceGuid(String apiEndpoint, String orgGuid, String space, String bearerToken, boolean verbose){
        def spaceInfo = script.httpRequest url: "${apiEndpoint}/v3/spaces?names=${space}&organization_guids=${orgGuid}", quiet: !verbose,
                                customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        def responseJson = script.readJSON text:"${spaceInfo.content}"
        if (!responseJson.resources[0]) throw new Exception("Cannot find Space")
        return responseJson.resources[0].guid
    }
}
