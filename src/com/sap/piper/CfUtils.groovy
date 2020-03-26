package com.sap.piper

def getXsuaaCredentials(String apiEndpoint, String org, String space, String credentialsId, String appName, boolean verbose){
    echo "Fetching xsuaa credentails of appName: ${appName} with parameters; apiEndpoint: ${apiEndpoint}, org: ${org}, space: ${space}, credentialsId: ${credentialsId}, verbose: ${verbose}"
    def responseJson
    withCredentials([usernamePassword(credentialsId: credentialsId, usernameVariable: 'usercf', passwordVariable: 'passwordcf')]) {
        //get authorization_endpoint
        def authorization_endpoint = httpRequest url: "${apiEndpoint}/v2/info", quiet: !verbose
        responseJson = readJSON text:"${authorization_endpoint.content}"

        //get token
        def access_token = httpRequest url:"${responseJson.authorization_endpoint}/oauth/token", quiet: !verbose,
                                httpMode:'POST',
                                requestBody: "username=${usercf}&password=${passwordcf}&client_id=cf&grant_type=password&response_type=token",
                                customHeaders: [[name: 'Content-Type', value: 'application/x-www-form-urlencoded'],[name: 'Authorization', value: 'Basic Y2Y6']]
        responseJson = readJSON text:"${access_token.content}"
        def bearerToken= "Bearer ${responseJson.access_token.trim()}"

        //get org guid
        def org_guid = httpRequest url: "${apiEndpoint}/v3/organizations?names=${org}", quiet: !verbose,
                                customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        responseJson = readJSON text:"${org_guid.content}"

        //get space guid
        def space_guid = httpRequest url: "${apiEndpoint}/v3/spaces?names=${space},organization_guids=${responseJson.resources[0].guid}", quiet: !verbose,
                                customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        responseJson = readJSON text:"${space_guid.content}"

        //get app guid
        def apps = httpRequest  url:"${apiEndpoint}/v3/apps?names=${appName},${appName}_blue,${appName}_green,space_guids=${responseJson.resources[0].guid}",  quiet: !verbose,
                                customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        responseJson = readJSON text:"${apps.content}"

        //get env variables
        def env = httpRequest   url: "${responseJson.resources[0].links.self.href.trim()}/env", quiet: !verbose,
                                customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
        responseJson = readJSON text:"${env.content}"
    }
    return responseJson.system_env_json.VCAP_SERVICES.xsuaa[0].credentials
}
