package com.sap.piper

import org.junit.Assert
import org.junit.Test

class BashUtilsTest {

    @Test
    void testEscape() {
        // input: a$b%c%d$e'$?$#$$"
        def input = "a\$b%c%d\$e\'\$?\$#\$\$\""
        // expect: 'a$b%c%d$e\'$?$#$$'
        def expected = "\'a\$b%c%d\$e\\\'\$?\$#\$\$\"\'"
        Assert.assertEquals(expected, BashUtils.escape(input))
    }

    @Test
    void testSimpleEscape() {
        // input: my file path
        def input = "my file path"
        // expect: 'my file path'
        def expected = "'my file path'"
        Assert.assertEquals(expected, BashUtils.escape(input))
    }
}
