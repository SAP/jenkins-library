package com.sap.piper.integration

def getXsuaaCredentials(String apiEndpoint, String org, String space, String credentialsId, String appName, boolean verbose){
    def env = getAppEnvironment(apiEndpoint, org, space, credentialsId, appName, verbose)
    return env.system_env_json.VCAP_SERVICES.xsuaa[0].credentials
}

def getAppEnvironment(String apiEndpoint, String org, String space, String credentialsId, String appName, boolean verbose){
    def auth_endpoint = getAuthEndPoint(apiEndpoint, verbose)
    def bearerToken = getBearerToken(auth_endpoint, credentialsId, verbose)
    def app_url = getAppRefUrl(apiEndpoint, org, space, bearerToken, appName, verbose)
    def env = httpRequest url: "${app_url}/env", quiet: !verbose,
                    customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
    def responseJson = readJSON text:"${env.content}"
    return responseJson
}

def getAuthEndPoint(String apiEndpoint, boolean verbose){
    def info = httpRequest url: "${apiEndpoint}/v2/info", quiet: !verbose
    def responseJson = readJSON text:"${info.content}"
    return responseJson.authorization_endpoint
}

def getBearerToken(String authorizationEndpoint, String credentialsId, boolean verbose){
    withCredentials([usernamePassword(credentialsId: credentialsId, usernameVariable: 'usercf', passwordVariable: 'passwordcf')]) {
        def token = httpRequest url:"${authorizationEndpoint}/oauth/token", quiet: !verbose,
                            httpMode:'POST',
                            requestBody: "username=${usercf}&password=${passwordcf}&client_id=cf&grant_type=password&response_type=token",
                            customHeaders: [[name: 'Content-Type', value: 'application/x-www-form-urlencoded'],[name: 'Authorization', value: 'Basic Y2Y6']]
        def responseJson = readJSON text:"${token.content}"
        return "Bearer ${responseJson.access_token.trim()}"
    }
}

def getAppRefUrl(String apiEndpoint, String org, String space, String bearerToken, String appName, boolean verbose){
    def org_guid = getOrgGuid(apiEndpoint, org, bearerToken, verbose)
    def space_guid = getSpaceGuid(apiEndpoint, org_guid, space, bearerToken, verbose)
    def appInfo = httpRequest  url:"${apiEndpoint}/v3/apps?names=${appName},${appName}_blue,${appName}_green,space_guids=${space_guid}",  quiet: !verbose,
                        customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
    def responseJson = readJSON text:"${appInfo.content}"
    return responseJson.resources[0].links.self.href.trim()
}

def getOrgGuid(String apiEndpoint, String org, String bearerToken, boolean verbose){
    def organizationInfo = httpRequest url: "${apiEndpoint}/v3/organizations?names=${org}", quiet: !verbose,
                                customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
    def responseJson = readJSON text:"${organizationInfo.content}"
    return responseJson.resources[0].guid
}

def getSpaceGuid(String apiEndpoint, String orgGuid, String space, String bearerToken, boolean verbose){
    def spaceInfo = httpRequest url: "${apiEndpoint}/v3/spaces?names=${space},organization_guids=${orgGuid}", quiet: !verbose,
                            customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
    def responseJson = readJSON text:"${spaceInfo.content}"
    return responseJson.resources[0].guid
}
