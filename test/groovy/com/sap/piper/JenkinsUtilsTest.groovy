package com.sap.piper

import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class JenkinsUtilsTest extends BasePiperTest {
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(shellRule)
        .around(loggingRule)

    JenkinsUtils jenkinsUtils
    Object currentBuildMock
    Object rawBuildMock
    Object jenkinsInstanceMock
    Object parentMock

    Map triggerCause

    String userId


    @Before
    void init() throws Exception {
        jenkinsUtils = new JenkinsUtils() {
            def getCurrentBuildInstance() {
                return currentBuildMock
            }

            def getActiveJenkinsInstance() {
                return jenkinsInstanceMock
            }
        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jenkinsUtils)

        jenkinsInstanceMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jenkinsInstanceMock)

        parentMock = new Object() {

        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(parentMock)

        rawBuildMock = new Object() {
            def getParent() {
                return parentMock
            }
            def getCause(type) {
                if (type == hudson.model.Cause.UserIdCause.class){
                    def userIdCause = new hudson.model.Cause.UserIdCause()
                    userIdCause.metaClass.getUserId =  {
                        return userId
                    }
                    return userIdCause
                } else {
                    return triggerCause
                }
            }

        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(rawBuildMock)

        currentBuildMock = new Object() {
            def number
            def getRawBuild() {
                return rawBuildMock
            }
        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(currentBuildMock)
    }
    @Test
    void testNodeAvailable() {
        def result = jenkinsUtils.nodeAvailable()
        assertThat(shellRule.shell, contains("echo 'Node is available!'"))
        assertThat(result, is(true))
    }

    @Test
    void testNoNodeAvailable() {
        helper.registerAllowedMethod('sh', [String.class], {s ->
            throw new MissingContextVariableException(String.class)
        })

        def result = jenkinsUtils.nodeAvailable()
        assertThat(loggingRule.log, containsString('No node context available.'))
        assertThat(result, is(false))
    }

    @Test
    void testGetIssueCommentTriggerAction() {
        triggerCause = [
            comment: 'this is my test comment /n /piper test whatever',
            triggerPattern: '.*/piper ([a-z]*).*'
        ]
        assertThat(jenkinsUtils.getIssueCommentTriggerAction(), is('test'))
    }

    @Test
    void testGetIssueCommentTriggerActionNoAction() {
        triggerCause = [
            comment: 'this is my test comment /n whatever',
            triggerPattern: '.*/piper ([a-z]*).*'
        ]
        assertThat(jenkinsUtils.getIssueCommentTriggerAction(), isEmptyOrNullString())
    }

    @Test
    void testGetUserId() {
        userId = 'Test User'
        assertThat(jenkinsUtils.getJobStartedByUserId(), is('Test User'))
    }

    @Test
    void testGetUserIdNoUser() {
        userId = null
        assertThat(jenkinsUtils.getJobStartedByUserId(), isEmptyOrNullString())
    }
}
