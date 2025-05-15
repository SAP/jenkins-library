import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Only if `notifyCulprits` is set:
     * Credentials if the repository is protected.
     * @possibleValues Jenkins credentials id
     */
    'gitSshKeyCredentialsId'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Set the build result used to determine the mail template.
     * default `currentBuild.result`
     */
    'buildResult',
    /**
     * Only if `notifyCulprits` is set:
     * Defines a dedicated git commitId for culprit retrieval.
     * default `commonPipelineEnvironment.getGitCommitId()`
     */
    'gitCommitId',
    /**
     * Only if `notifyCulprits` is set:
     * Repository url used to retrieve culprit information.
     * default `commonPipelineEnvironment.getGitSshUrl()`
     */
    'gitUrl',
    /**
     * defines if the console log file should be attached to the notification mail.
     * @possibleValues `true`, `false`
     */
    'notificationAttachment',
    /**
     * A space-separated list of recipients that always get the notification.
     */
    'notificationRecipients',
    /**
     * Notify all committers since the last successful build.
     * @possibleValues `true`, `false`
     */
    'notifyCulprits',
    /**
     * Number of log line which are included in the email body.
     */
    'numLogLinesInBody',
    /**
     * The project name used in the email subject.
     * default `currentBuild.fullProjectName`
     */
    'projectName',
    /**
     * Needs to be set to `true` if step is used outside of a node context, e.g. post actions in a declarative pipeline script.
     * default `false`
     * @possibleValues `true`, `false`
     */
    'wrapInNode'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Sends notifications to all potential culprits of a current or previous build failure and to fixed list of recipients.
 * It will attach the current build log to the email.
 *
 * Notifications are sent in following cases:
 *
 * * current build failed or is unstable
 * * current build is successful and previous build failed or was unstable
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {
        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(
                projectName: script.currentBuild.fullProjectName,
                displayName: script.currentBuild.displayName,
                buildResult: script.currentBuild.result,
                gitUrl: script.commonPipelineEnvironment.getGitSshUrl(),
                gitCommitId: script.commonPipelineEnvironment.getGitCommitId()
            )
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        //this takes care that terminated builds due to milestone-locking do not cause an error
        if (script.commonPipelineEnvironment.getBuildResult() == 'ABORTED') return

        def subject = "${config.buildResult}: Build ${config.projectName} ${config.displayName}"
        def log = ''
        def mailTemplate
        if (config.buildResult == 'UNSTABLE' || config.buildResult == 'FAILURE'){
            mailTemplate = 'com.sap.piper/templates/mailFailure.html'
            log = script.currentBuild.rawBuild.getLog(config.numLogLinesInBody).join('\n')
        }else if(hasRecovered(config.buildResult, script.currentBuild)){
            mailTemplate = 'com.sap.piper/templates/mailRecover.html'
            subject += ' is back to normal'
        }
        if(mailTemplate){
            def mailContent = GStringTemplateEngine.newInstance().createTemplate(libraryResource(mailTemplate)).make([env: env, log: log]).toString()
            def recipientList = ''
            if(config.notifyCulprits){
                if (!config.gitUrl) {
                    echo "[${STEP_NAME}] no gitUrl available, -> exiting without sending mails"
                    return
                }
                recipientList += getCulpritCommitters(config, script.currentBuild)
            }
            if(config.notificationRecipients)
                recipientList +=  " ${config.notificationRecipients}"
            emailext(
                mimeType: 'text/html',
                subject: subject.trim(),
                body: mailContent,
                to: recipientList.trim(),
                recipientProviders: [requestor()],
                attachLog: config.notificationAttachment
            )
        }
    }
}

def getNumberOfCommits(buildList){
    def numCommits = 0
    if(buildList != null)
        for(actBuild in buildList) {
            def changeLogSets = actBuild.getChangeSets()
            if(changeLogSets != null)
                for(changeLogSet in changeLogSets)
                    for(change in changeLogSet)
                        numCommits++
        }
    return numCommits
}

def getCulpritCommitters(config, currentBuild) {
    def recipients
    def buildList = []
    def build = currentBuild

    if (build != null) {
        // At least add the current build
        buildList.add(build)

        // Now collect FAILED or ABORTED ones
        build = build.getPreviousBuild()
        while (build != null) {
            if (build.getResult() != 'SUCCESS') {
                buildList.add(build)
            } else {
                break
            }
            build = build.getPreviousBuild()
        }
    }
    def numberOfCommits = getNumberOfCommits(buildList)
    if(config.wrapInNode){
        node(){
            try{
                recipients = getCulprits(config, env.BRANCH_NAME, numberOfCommits)
            }finally{
                deleteDir()
            }
        }
    }else{
        try{
            recipients = getCulprits(config, env.BRANCH_NAME, numberOfCommits)
        }finally{
            deleteDir()
        }
    }
    echo "[${STEP_NAME}] last ${numberOfCommits} commits revealed following responsibles ${recipients}"
    return recipients
}

def getCulprits(config, branch, numberOfCommits) {
    try {
        if (branch?.startsWith('PR-')) {
            //special GitHub Pull Request handling
            deleteDir()
            sshagent(
                credentials: [config.gitSshKeyCredentialsId],
                ignoreMissing: true
            ) {
                def pullRequestID = branch.replaceAll('PR-', '')
                def localBranchName = "pr" + pullRequestID
                sh """git init
    git fetch ${q(config.gitUrl)} pull/${q(pullRequestID)}/head:${q(localBranchName)} > /dev/null 2>&1
    git checkout -f ${q(localBranchName)} > /dev/null 2>&1
    """
            }
        } else {
            //standard git/GitHub handling
            if (config.gitCommitId) {
                deleteDir()
                sshagent(
                    credentials: [config.gitSshKeyCredentialsId],
                    ignoreMissing: true
                ) {
                    sh """git clone ${q(config.gitUrl)} .
    git checkout ${q(config.gitCommitId)} > /dev/null 2>&1"""
                }
            } else {
                def retCode = sh(returnStatus: true, script: 'git log > /dev/null 2>&1')
                if (retCode != 0) {
                    echo "[${STEP_NAME}] No git context available to retrieve culprits"
                    return ''
                }
            }
        }

        def recipients = sh(returnStdout: true, script: "git log -${numberOfCommits} --first-parent --pretty=format:'%ae %ce'")
        return getDistinctRecipients(recipients)
    } catch(err) {
        echo "[${STEP_NAME}] Culprit retrieval from git failed with '${err.getMessage()}'. Please make sure to configure gitSshKeyCredentialsId. So far, only fixed list of recipients is used."
        return ''
    }
}

def getDistinctRecipients(recipients){
    def result
    def recipientAddresses = recipients.split()
    def knownAddresses = new HashSet<String>()
    if(recipientAddresses != null) {
        for(address in recipientAddresses) {
            address = address.trim()
            if(address
                && address.contains("@")
                && !address.contains("noreply")
                && !knownAddresses.contains(address)) {
                knownAddresses.add(address)
            }
        }
        result = knownAddresses.join(" ")
    }
    return result
}

def hasRecovered(buildResult, currentBuild){
    def previousBuild = currentBuild.getPreviousBuild()

    if (previousBuild) {
        return buildResult == 'SUCCESS' && previousBuild.result != 'SUCCESS'
    }

    return false
}
