package com.sap.piper

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsReadMavenPomRule
import util.JenkinsShellCallRule
import util.Rules
import util.SharedLibraryCreator

import static org.junit.Assert.assertEquals

class GitUtilsTest extends BasePipelineTest {

    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this).around(jscr).around(thrown)

    GitUtils gitUtils

    @Before
    void init() throws Exception {
        gitUtils = new GitUtils()
        prepareObjectInterceptors(gitUtils)

        jscr.setReturnValue('git rev-parse HEAD', 'testCommitId')
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }

    @Test
    void testGetGitCommitId() {

        assertEquals('testCommitId', gitUtils.getGitCommitId())

    }

}
