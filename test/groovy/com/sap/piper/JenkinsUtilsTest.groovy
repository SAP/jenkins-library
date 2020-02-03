package com.sap.piper

import hudson.AbortException
import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class JenkinsUtilsTest extends BasePiperTest {
    public ExpectedException exception = ExpectedException.none()

    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(exception)
        .around(shellRule)
        .around(loggingRule)

    JenkinsUtils jenkinsUtils
    Object currentBuildMock
    Object rawBuildMock
    Object jenkinsInstanceMock
    Object parentMock

    Map triggerCause

    String userId

    Map results


    @Before
    void init() throws Exception {
        results = [:]
        results.runlinkCalled = false
        results.joblinkCalled = false
        results.removejoblinkCalled = false

        jenkinsUtils = new JenkinsUtils() {
            def getCurrentBuildInstance() {
                return currentBuildMock
            }

            def getActiveJenkinsInstance() {
                return jenkinsInstanceMock
            }

            void addRunSideBarLink(String relativeUrl, String displayName, String relativeIconPath) {
                results.runlinkCalled = true
                assertThat(relativeUrl, is('https://server.com/1234.pdf'))
                assertThat(displayName, is('Test link'))
                assertThat(relativeIconPath, is('images/24x24/graph.png'))
            }

            void addJobSideBarLink(String relativeUrl, String displayName, String relativeIconPath) {
                results.joblinkCalled = true
                assertThat(relativeUrl, is('https://server.com/1234.pdf'))
                assertThat(displayName, is('Test link'))
                assertThat(relativeIconPath, is('images/24x24/graph.png'))
            }
            void removeJobSideBarLinks(String relativeUrl) {
                results.removejoblinkCalled = true
                assertThat(relativeUrl, is('https://server.com/1234.pdf'))
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
            def getAction(type) {
                return new Object() {
                    def getLibraries() {
                        return [
                            [name: 'lib1', version: '1', trusted: true],
                            [name: 'lib2', version: '2', trusted: false],
                        ]
                    }
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
    void testHandleStepResultsJobLink() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return []
            if(m.file == 'someStep_links.json')
                return [[target: "https://server.com/1234.pdf", name: "Test link", mandatory: true, scope: 'job']]
        })

        jenkinsUtils.handleStepResults("someStep", true, true)

        assertThat(results.removejoblinkCalled, is(true))
        assertThat(results.runlinkCalled, is(true))
        assertThat(results.joblinkCalled, is(true))
    }
    @Test
    void testHandleStepResults() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'someStep_links.json')
                return [[target: "https://server.com/1234.pdf", name: "Test link", mandatory: true]]
        })

        jenkinsUtils.handleStepResults("someStep", true, true)

        assertThat(results.removejoblinkCalled, is(false))
        assertThat(results.runlinkCalled, is(true))
        assertThat(results.joblinkCalled, is(false))
    }
    @Test
    void testHandleStepResultsEmptyReports() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return []
            if(m.file == 'someStep_links.json')
                return [[target: "https://server.com/1234.pdf", name: "Test link", mandatory: true]]
        })

        jenkinsUtils.handleStepResults("someStep", true, true)
    }
    @Test
    void testHandleStepResultsEmptyLinks() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'someStep_links.json')
                return []
        })

        jenkinsUtils.handleStepResults("someStep", true, true)
    }
    @Test
    void testHandleStepResultsNoErrorReportsLinks() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return true
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return []
            if(m.file == 'someStep_links.json')
                return []
        })
        jenkinsUtils.handleStepResults("someStep", false, false)
    }
    @Test
    void testHandleStepResultsReportsNoFile() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return false
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'someStep_links.json')
                return [[target: "https://server.com/1234.pdf", name: "Test link", mandatory: true]]
        })

        exception.expect(AbortException)
        exception.expectMessage("Expected to find someStep_reports.json in workspace but it is not there")

        jenkinsUtils.handleStepResults("someStep", true, false)
    }
    @Test
    void testHandleStepResultsLinksNoFile() {
        helper.registerAllowedMethod("fileExists", [Map], { m ->
            return false
        })
        helper.registerAllowedMethod("readJSON", [Map], { m ->
            if(m.file == 'someStep_reports.json')
                return [[target: "1234.pdf", mandatory: true]]
            if(m.file == 'someStep_links.json')
                return [[target: "https://server.com/1234.pdf", name: "Test link", mandatory: true]]
        })
        helper.registerAllowedMethod('addRunSideBarLink', [String, String, String], { u, n, i ->
            assertThat(u, is('https://server.com/1234.pdf'))
            assertThat(n, is('Test link'))
            assertThat(i, is('images/24x24/graph.png'))
        })

        exception.expect(AbortException)
        exception.expectMessage("Expected to find someStep_links.json in workspace but it is not there")

        jenkinsUtils.handleStepResults("someStep", false, true)
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

    @Test
    void testGetLibrariesInfo() {
        def libs
        libs = jenkinsUtils.getLibrariesInfo()
        assertThat(libs[0], is([name: 'lib1', version: '1', trusted: true]))
        assertThat(libs[1], is([name: 'lib2', version: '2', trusted: false]))
    }
}
