package com.sap.piper

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.Rules

import java.nio.file.Paths

import static org.junit.Assert.assertEquals

class PathUtilsTest extends BasePiperTest {

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)


    @Test
    void testConvertToAbsolutePath() {
        nullScript.env.WORKSPACE = '/workspace'

        String result = PathUtils.convertToAbsolutePath(nullScript, 'test')

        assertEquals('/workspace/test', result)
    }
}
