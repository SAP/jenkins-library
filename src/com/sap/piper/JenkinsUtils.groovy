package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import hudson.tasks.junit.TestResultAction

import jenkins.model.Jenkins

import org.apache.commons.io.IOUtils
import org.jenkinsci.plugins.workflow.libs.LibrariesAction
import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException

@API
@NonCPS
static def isPluginActive(pluginId) {
    return Jenkins.instance.pluginManager.plugins.find { p -> p.isActive() && p.getShortName() == pluginId }
}

static boolean hasTestFailures(build){
    //build: https://javadoc.jenkins.io/plugin/workflow-support/org/jenkinsci/plugins/workflow/support/steps/build/RunWrapper.html
    //getRawBuild: https://javadoc.jenkins.io/plugin/workflow-job/org/jenkinsci/plugins/workflow/job/WorkflowRun.html
    //getAction: http://www.hudson-ci.org/javadoc/hudson/tasks/junit/TestResultAction.html
    def action = build?.getRawBuild()?.getAction(TestResultAction.class)
    return action && action.getFailCount() != 0
}

boolean addWarningsNGParser(String id, String name, String regex, String script, String example = ''){
    def classLoader = this.getClass().getClassLoader()
    // usage of class loader to avoid plugin dependency for other use cases of JenkinsUtils class
    def parserConfig = classLoader.loadClass('io.jenkins.plugins.analysis.warnings.groovy.ParserConfiguration', true)?.getInstance()

    if(parserConfig.contains(id)){
        return false
    }else{
        parserConfig.setParsers(
            parserConfig.getParsers().plus(
                classLoader.loadClass('io.jenkins.plugins.analysis.warnings.groovy.GroovyParser', true)?.newInstance(id, name, regex, script, example)
            )
        )
        return true
    }
}

@NonCPS
static String getFullBuildLog(currentBuild) {
    Reader reader = currentBuild.getRawBuild().getLogReader()
    String logContent = IOUtils.toString(reader);
    reader.close();
    reader = null
    return logContent
}

def nodeAvailable() {
    try {
        sh "echo 'Node is available!'"
    } catch (MissingContextVariableException e) {
        echo "No node context available."
        return false
    }
    return true
}


@NonCPS
def getCurrentBuildInstance() {
    return currentBuild
}

@NonCPS
def getParentJob() {
    return getRawBuild().getParent()
}

@NonCPS
def getRawBuild() {
    return getCurrentBuildInstance().rawBuild
}

def isJobStartedByTimer() {
    return isJobStartedByCause(hudson.triggers.TimerTrigger.TimerTriggerCause.class)
}

def isJobStartedByUser() {
    return isJobStartedByCause(hudson.model.Cause.UserIdCause.class)
}

@NonCPS
def isJobStartedByCause(Class cause) {
    def startedByGivenCause = false
    def detectedCause = getRawBuild().getCause(cause)
    if (null != detectedCause) {
        startedByGivenCause = true
        echo "Found build cause ${detectedCause}"
    }
    return startedByGivenCause
}

@NonCPS
String getIssueCommentTriggerAction() {
    try {
        def triggerCause = getRawBuild().getCause(org.jenkinsci.plugins.pipeline.github.trigger.IssueCommentCause)
        if (triggerCause) {
            //triggerPattern e.g. like '.* /piper ([a-z]*) .*'
            def matcher = triggerCause.comment =~ triggerCause.triggerPattern
            if (matcher) {
                return matcher[0][1]
            }
        }
        return null
    } catch (err) {
        return null
    }
}

def getJobStartedByUserId() {
    return getRawBuild().getCause(hudson.model.Cause.UserIdCause.class)?.getUserId()
}

@NonCPS
def getLibrariesInfo() {
    def libraries = []
    def build = getRawBuild()
    def libs = build.getAction(LibrariesAction.class).getLibraries()

    for (def i = 0; i < libs.size(); i++) {
        Map lib = [:]

        lib['name'] = libs[i].name
        lib['version'] = libs[i].version
        lib['trusted'] = libs[i].trusted
        libraries.add(lib)
    }

    return libraries
}

@NonCPS
void addRunSideBarLink(String relativeUrl, String displayName, String relativeIconPath) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        if (relativeUrl != null && displayName != null) {
            def run = getRawBuild()
            def iconPath = (null != relativeIconPath) ? "${Functions.getResourcePath()}/${relativeIconPath}" : null
            def action = linkActionClass.newInstance(relativeUrl, displayName, iconPath)
            echo "Added run level sidebar link to '${action.getUrlName()}' with name '${action.getDisplayName()}' and icon '${action.getIconFileName()}'"
            run.getActions().add(action)
        }
    } catch (e) {
        e.printStackTrace()
    }
}

@NonCPS
void addJobSideBarLink(String relativeUrl, String displayName, String relativeIconPath) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        if (relativeUrl != null && displayName != null) {
            def parentJob = getParentJob()
            def buildNumber = getCurrentBuildInstance().number
            def iconPath = (null != relativeIconPath) ? "${Functions.getResourcePath()}/${relativeIconPath}" : null
            def action = linkActionClass.newInstance("${buildNumber}/${relativeUrl}", displayName, iconPath)
            echo "Added job level sidebar link to '${action.getUrlName()}' with name '${action.getDisplayName()}' and icon '${action.getIconFileName()}'"
            parentJob.getActions().add(action)
        }
    } catch (e) {
        e.printStackTrace()
    }
}

@NonCPS
void removeJobSideBarLinks(String relativeUrl = null) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        def parentJob = getParentJob()
        def listToRemove = new ArrayList()
        for (def action : parentJob.getActions()) {
            if (linkActionClass.isAssignableFrom(action.getClass()) && (null == relativeUrl || action.getUrlName().endsWith(relativeUrl))) {
                echo "Removing job level sidebar link to '${action.getUrlName()}' with name '${action.getDisplayName()}' and icon '${action.getIconFileName()}'"
                listToRemove.add(action)
            }
        }
        parentJob.getActions().removeAll(listToRemove)
        echo "Removed Jenkins global sidebar links ${listToRemove}"
    } catch (e) {
        e.printStackTrace()
    }
}

def getInstance() {
    Jenkins.get()
}
