package com.sap.piper

class BashUtils implements Serializable {
    static final long serialVersionUID = 1L
    public static final String ESCAPED_SINGLE_QUOTE = "'\"'\"'"

    /**
     * Put string in single quotes and escape contained single quotes by putting them into a double quoted string
     */
    static String quoteAndEscape(String str) {
        if(str == null) {
            return 'null'
        }
        def escapedString = str.replace("'", ESCAPED_SINGLE_QUOTE)
        return "'${escapedString}'"
    }
}
