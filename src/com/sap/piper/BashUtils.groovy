package com.sap.piper

class BashUtils implements Serializable {
    static final long serialVersionUID = 1L

    /**
     * Put string in single quotes and escape contained single quotes by putting them into a double quoted string
     */
    static String quoteAndEscape(String str) {
        def escapedString = str.replace("'", "'\"'\"'")
        return "'${escapedString}'"
    }

    /**
     * Checks if string is quoted in single quotes, and if so, removes them and replaces and occurrences of
     * '"'"' with '
     */
    static String unQuoteAndEscape(String str) {
        if (str.startsWith("'") && str.endsWith("'")) {
            str = str.substring(1, str.length() - 1)
            String test = str.replace("'\"'\"'", "")
            if (test.contains("'")) {
                throw new IllegalArgumentException('String is quoted, but contains unescaped single quotes')
            }
            str = str.replace("'\"'\"'", "'")
        }
        return str
    }
}
