import com.sap.piper.SidecarUtils


import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.JsonUtils

import groovy.transform.Field
import hudson.AbortException
import groovy.json.JsonSlurper;

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Specifies the host address
     */
    'host',
    /**
     * Specifies the name of the Software Component
     */
    'name',
    /**
     * Specifies the communication user of the communication scenario SAP_COM_0510
     */
    'username',
    /**
     * Specifies the password of the communication user
     */
    'password'])
/**
 * Pulls a Software Component to a SAP Cloud Platform ABAP Environment System.
 *
 * Prerequisite: the Communication Arrangement for the Communication Scenario SAP_COM_0510 has to be set up, including a Communication System and Communication Arrangement
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        String host = parameters.host;
        String repositoryName = parameters.repositoryName;
        String username = parameters.username;
        String password = parameters.password;

        String usernameColonPassword = username + ":" + password;
        String authToken = usernameColonPassword.bytes.encodeBase64().toString()
        String port = ':443';
        String service = '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY';
        String entity = '/Pull';
        String urlString = host + port + service + entity;
        println "API: " + urlString;

        def url = new URL(urlString);
        Map tokenAndCookie = getTokenAndCookie(url, authToken);
        String token = tokenAndCookie.token;
        String cookie = tokenAndCookie.cookie;

        HttpURLConnection connection = createPostConnection(url, token, cookie, authToken)
        connection.connect();
        OutputStream outputStream = connection.getOutputStream();
        String input = '{ "sc_name" : "' + repositoryName + '" }';
        println input;
        outputStream.write(input.getBytes());
        outputStream.flush();

        int statusCode = connection.responseCode;

        if (statusCode == 200 || statusCode == 201) {

            String body = connection.content.text;

            JsonSlurper slurper = new JsonSlurper();
            Map object = slurper.parseText(body);
            connection.disconnect();
            println object.d."status_descr";
            String pollUri = object.d."__metadata"."uri"
            println pollUri;
            def pollUrl = new URL(pollUri);

            while({
                Thread.sleep(5000);
                HttpURLConnection pollConnection = createDefaultConnection(pollUrl, authToken);
                pollConnection.connect();
                int pollStatusCode = pollConnection.responseCode;
                if (pollStatusCode == 200 || pollStatusCode == 201) {
                    String pollBody = pollConnection.content.text;
                    Map pollObject = slurper.parseText(pollBody);
                    String pollStatus = pollObject.d."status";
                    String pollStatusText = pollObject.d."status_descr";
                    pollConnection.disconnect();
                    if (pollStatus == 'R') {
                        true;
                    } else {
                        println pollStatusText;
                        if (pollStatus != 'S') {
                            throw new Exception("Pull Failed");
                        }
                        false;
                    }
                } else {
                    println pollConnection.getErrorStream().text;
                    pollConnection.disconnect();
                    throw new Exception("HTTPS Connection Failed");
                    false;
                }

            }()) continue
            
        } else {
            println connection.getErrorStream().text;
            connection.disconnect();
            throw new Exception("HTTPS Connection Failed");
        }
    }
}


def Map getTokenAndCookie(URL url, String authToken) {
    HttpURLConnection connection = createDefaultConnection(url, authToken);
    connection.setRequestProperty("x-csrf-token", "fetch");

    System.setProperty("https.protocols", "TLSv1,TLSv1.1,TLSv1.2");

    connection.setRequestMethod("GET");
    connection.connect();
    token =  connection.getHeaderField("x-csrf-token");
    cookie1 = connection.getHeaderField(1).split(";")[0] 
    cookie2 = connection.getHeaderField(2).split(";")[0] 
    cookie = cookie1 + "; " + cookie2; 
    connection.disconnect();
    connection = null;

    Map result = [:];
    result.cookie = cookie;
    result.token = token;
    return result;

}

def HttpURLConnection createDefaultConnection(URL url, String authToken) {
    HttpURLConnection connection = (HttpURLConnection) url.openConnection();
    connection.setRequestProperty("Authorization", "Basic " + authToken);
    connection.setRequestProperty("Content-Type", "application/json");
    connection.setRequestProperty("Accept", "application/json");
    return connection;

}

def HttpURLConnection createPostConnection(URL url, String token, String cookie, String authToken) {

    HttpURLConnection connection = createDefaultConnection(url, authToken);
    connection.setRequestProperty("cookie", cookie);
    connection.setRequestProperty("x-csrf-token", token);
    connection.setRequestMethod("POST");
    connection.setDoOutput(true);
    connection.setDoInput(true);
    return connection;

}