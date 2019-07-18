package com.sap.piper

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.Rules

import static org.junit.Assert.assertEquals

class PathUtilsTest extends BasePiperTest {

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)

    @Test
    void testFillPathTemplate() {
        nullScript.env.WORKSPACE = '/workspace'
        String result = PathUtils.fillPathTemplate(nullScript, '${workspaceRoot}/test')
        assertEquals('/workspace/test', result)
    }

    @Test
    void testReplacePathInConfiguration() {
        nullScript.env.WORKSPACE = '/workspace'

        Map configuration = [key1:null, key2:1, key3:'${workspaceRoot}/test']

        Map result = PathUtils.replacePathInConfiguration(nullScript, configuration, ['key1', 'key3'] as Set)
        assertEquals('/workspace/test', result.key3)
    }
}
