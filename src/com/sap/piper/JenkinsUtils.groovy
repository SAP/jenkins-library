package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import jenkins.model.Jenkins
import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException
import hudson.tasks.junit.TestResultAction

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
