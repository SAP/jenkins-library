package com.sap.piper

import org.junit.Assert
import org.junit.Test

class BashUtilsTest {
    @Test
    void escapeFilePath() {
        // Given: A Windows-style file path C:\some\path
        def input = 'C:\\some\\path'

        // When we escape the string
        def result = BashUtils.quoteAndEscape(input)

        // Then the string is surrounded by single quotes 'C:\some\path'
        def expected = "'C:\\some\\path'"
        Assert.assertEquals(expected, result)
    }

    @Test
    void escapeUri() {
        // Given: An URI with single quotes values http://www.sap.com?$filter='234'
        def input = "http://www.sap.com?\$filter='234'"

        // When we escape the string
        def result = BashUtils.quoteAndEscape(input)

        // Then the input string is surrounded by single quotes and each original ' is replaced by '"'"'
        // 'http://www.sap.com?$filter='"'"'234'"'"''
        def expected = "'http://www.sap.com?\$filter='\"'\"'234'\"'\"''"
        Assert.assertEquals(expected, result)
    }

    @Test
    void escapePassword() {
        // Given: A random generated password VQ5r\%*h"49'Ch>Jj?
        def input = "VQ5r\\%*h\"49'Ch>Jj?"

        // When we escape the string
        def result = BashUtils.quoteAndEscape(input)

        // Then the input string is surrounded by single quotes and each original ' is replaced by '"'"'
        // 'VQ5r\%*h"49'"'"'Ch>Jj?'
        def expected = "'VQ5r\\%*h\"49'\"'\"'Ch>Jj?'"
        Assert.assertEquals(expected, result)
    }

    @Test
    void unEscapeEmpty() {
        def input = ""
        def result = BashUtils.unQuoteAndEscape(input)
        Assert.assertEquals(input, result)
    }

    @Test
    void unEscapeSingleQuote() {
        def input = "'"
        def result = BashUtils.unQuoteAndEscape(input)
        Assert.assertEquals(input, result)
    }

    @Test
    void unEscapeUnquoted() {
        def input = "some '\"'\"' test"
        def result = BashUtils.unQuoteAndEscape(input)
        Assert.assertEquals(input, result)
    }

    @Test
    void unEscapeStartsWithQuote() {
        def input = "'some '\"'\"' test"
        def result = BashUtils.unQuoteAndEscape(input)
        Assert.assertEquals(input, result)
    }

    @Test
    void unEscapeEndsWithQuote() {
        def input = "some '\"'\"' test'"
        def result = BashUtils.unQuoteAndEscape(input)
        Assert.assertEquals(input, result)
    }

    @Test
    void unEscapeQuoted() {
        def input = "''\"'\"'some '\"'\"' test'\"'\"''"
        def result = BashUtils.unQuoteAndEscape(input)
        Assert.assertEquals("'some ' test'", result)
    }

    @Test
    void unEscapeInvalid() {
        def input = "'''"
        boolean exceptionThrown = false
        try {
            BashUtils.unQuoteAndEscape(input)
        } catch (Exception ignored) {
            exceptionThrown = true
        }
        Assert.assertTrue(exceptionThrown)
    }
}
