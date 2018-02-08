package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@NonCPS
def getMandatoryParameter(Map map, paramName, defaultValue = null) {

    def paramValue = map[paramName]

    if (paramValue == null)
        paramValue = defaultValue

    if (paramValue == null)
        throw new Exception("ERROR - NO VALUE AVAILABLE FOR ${paramName}")
    return paramValue

}

def getMtaJar(script, stepName, configuration, environment) {
    def mtaJarLocation = 'mta.jar' //default, maybe it is in current working directory

    if(configuration?.mtaJarLocation){
        mtaJarLocation = "${configuration.mtaJarLocation}/mta.jar"
        script.echo "[$stepName] MTA JAR \"${mtaJarLocation}\" retrieved from configuration."
        return mtaJarLocation
    }

    if(environment?.MTA_JAR_LOCATION){
        mtaJarLocation = "${environment.MTA_JAR_LOCATION}/mta.jar"
        script.echo "[$stepName] MTA JAR \"${mtaJarLocation}\" retrieved from environment."
        return mtaJarLocation
    }

    script.echo "[$stepName] Using MTA JAR from current working directory."
    return mtaJarLocation
}

def getNeoExecutable(script, stepName, configuration, environment) {

    def neoExecutable = 'neo.sh' // default, if nothing below applies maybe it is the path.

    if (configuration?.neoHome) {
        neoExecutable = "${configuration.neoHome}/tools/neo.sh"
        script.echo "[$stepName] Neo executable \"${neoExecutable}\" retrieved from configuration."
        return neoExecutable
    }

    if (environment?.NEO_HOME) {
        neoExecutable = "${environment.NEO_HOME}/tools/neo.sh"
        script.echo "[$stepName] Neo executable \"${neoExecutable}\" retrieved from environment."
        return neoExecutable
    }

    script.echo "[$stepName] Using Neo executable from PATH."
    return neoExecutable
}

def getCmCliExecutable(script, stepName, configuration, environment) {

    def cmCliExecutable = 'cmclient' // default, if nothing below applies maybe it is the path.

    if (configuration?.cmCliHome) {
        cmCliExecutable = "${configuration.cmCliHome}/bin/cmclient"
        script.echo "[$stepName] Change Management Command Line Interface \"${cmCliExecutable}\" retrieved from configuration."
        return cmCliExecutable
    }

    if (environment?.CM_CLI_HOME) {
        cmCliExecutable = "${environment.CM_CLI_HOME}/bin/cmclient"
        script.echo "[$stepName] Change Management Command Line Interface \"${cmCliExecutable}\" retrieved from environment."
        return cmCliExecutable
    }

    script.echo "[$stepName] Change Management Command Line Interface retrieved from current working directory."
    return cmCliExecutable
}

