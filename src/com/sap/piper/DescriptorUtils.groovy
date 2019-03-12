package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import groovy.transform.Field

import java.util.regex.Matcher
import java.util.regex.Pattern

@Field
def name = Pattern.compile("(.*)name=['\"](.*?)['\"](.*)", Pattern.DOTALL)
@Field
def version = Pattern.compile("(.*)version=['\"](.*?)['\"](.*)", Pattern.DOTALL)
@Field
def method = Pattern.compile("(.*)\\(\\)", Pattern.DOTALL)

def getMavenGAV(fileName) {
    def result = [:]
    def descriptor = readMavenPom(file: fileName)
    def group = descriptor.getGroupId()
    def artifact = descriptor.getArtifactId()
    def version = descriptor.getVersion()
    result['packaging'] = descriptor.getPackaging()
    result['group'] = (null != group && group.length() > 0) ? group : sh(returnStdout: true, script: "mvn -f ${fileName} help:evaluate -Dexpression=project.groupId | grep -Ev '(^\\s*\\[|Download|Java\\w+:)'").trim()
    result['artifact'] = (null != artifact && artifact.length() > 0) ? artifact : sh(returnStdout: true, script: "mvn -f ${fileName} help:evaluate -Dexpression=project.artifactId | grep -Ev '(^\\s*\\[|Download|Java\\w+:)'").trim()
    result['version'] = (null != version && version.length() > 0) ? version : sh(returnStdout: true, script: "mvn -f ${fileName} help:evaluate -Dexpression=project.version | grep ^[0-9].*").trim()
    echo "loaded ${result} from ${fileName}"
    return result
}

def getNpmGAV(file = 'package.json') {
    def result = [:]
    def descriptor = readJSON(file: file)

    if (descriptor.name.startsWith('@')) {
        def packageNameArray = descriptor.name.split('/')
        if (packageNameArray.length != 2)
            error "Unable to parse package name '${descriptor.name}'"
        result['group'] = packageNameArray[0]
        result['artifact'] = packageNameArray[1]
    } else {
        result['group'] = ''
        result['artifact'] = descriptor.name
    }
    result['version'] = descriptor.version
    echo "loaded ${result} from ${file}"
    return result
}

def getDlangGAV(file = 'dub.json') {
    def result = [:]
    def descriptor = readJSON(file: file)

    result['group'] = 'com.sap.dlang'
    result['artifact'] = descriptor.name
    result['version'] = descriptor.version
    result['packaging'] = 'tar.gz'
    echo "loaded ${result} from ${file}"
    return result
}

def getSbtGAV(file = 'sbtDescriptor.json') {
    def result = [:]
    def descriptor = readJSON(file: file)

    result['group'] = descriptor.group
    result['artifact'] = descriptor.artifactId
    result['version'] = descriptor.version
    result['packaging'] = descriptor.packaging
    echo "loaded ${result} from ${file}"
    return result
}

def getPipGAV(file = 'setup.py') {
    def result = [:]
    def descriptor = sh(returnStdout: true, script: "cat ${file}")

    result['group'] = ''
    result['packaging'] = ''
    result['artifact'] = matches(name, descriptor)
    result['version'] = matches(version, descriptor)

    if (result['version'] == '' || matches(method, result['version'])) {
        file = file.replace('setup.py', 'version.txt')
        def versionString = sh(returnStdout: true, script: "cat ${file}")
        if (versionString) {
            result['version'] = versionString.trim()
        }
    }

    echo "loaded ${result} from ${file}"
    return result
}

@NonCPS
private def matches(regex, input) {
    def m = new Matcher(regex, input)
    return m.matches() ? m.group(2) : ''
}
