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

def getMavenGAV(file = 'pom.xml') {
    def result = [:]
    def descriptor = readMavenPom(file: file)
    def group = descriptor.getGroupId()
    def artifact = descriptor.getArtifactId()
    def version = descriptor.getVersion()
    result['packaging'] = descriptor.getPackaging()
    result['group'] = (null != group && group.length() > 0) ? group : sh(returnStdout: true, script: "mvn -f ${file} help:evaluate -Dexpression=project.groupId | grep -Ev '(^\\s*\\[|Download|Java\\w+:)'").trim()
    result['artifact'] = (null != artifact && artifact.length() > 0) ? artifact : sh(returnStdout: true, script: "mvn -f ${file} help:evaluate -Dexpression=project.artifactId | grep -Ev '(^\\s*\\[|Download|Java\\w+:)'").trim()
    result['version'] = (null != version && version.length() > 0) ? version : sh(returnStdout: true, script: "mvn -f ${file} help:evaluate -Dexpression=project.version | grep ^[0-9].*").trim()
    echo "loaded ${result} from ${file}"
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

def getDubGAV(file = 'dub.json') {
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
    def descriptor = readFile(file: file)

    result['group'] = ''
    result['packaging'] = ''
    result['artifact'] = matches(name, descriptor)
    result['version'] = matches(version, descriptor)

    if (result['version'] == '' || matches(method, result['version'])) {
        file = file.replace('setup.py', 'version.txt')
        result['version'] = getVersionFromFile(file)
    }

    echo "loaded ${result} from ${file}"
    return result
}

def getGoGAV(file = 'Gopkg.toml', URI repoUrl) {
    def name = "${repoUrl.getHost()}${repoUrl.getPath().replaceAll(/\.git/, '')}"
    def path = file.substring(0, file.lastIndexOf('/') + 1)
    def module = path?.replaceAll(/\./, '')?.replaceAll('/', '')
    def result = [:]

    result['group'] = ''
    result['packaging'] = ''
    result['artifact'] = "${name}${module?'.':''}${module?:''}".toString()
    file = path + 'version.txt'
    result['version'] = getVersionFromFile(file)

    if (!result['version']) {
        file = path + 'VERSION'
        result['version'] = getVersionFromFile(file)
    }

    echo "loaded ${result} from ${file}"
    return result
}

private getVersionFromFile(file) {
    try {
        def versionString = readFile(file: file)
        if (versionString) {
            return versionString.trim()
        }
    } catch (java.nio.file.NoSuchFileException e) {
        echo "Failed to load version string from file ${file} due to ${e}"
    }
    return ''
}

@NonCPS
private def matches(regex, input) {
    def m = new Matcher(regex, input)
    return m.matches() ? m.group(2) : ''
}
