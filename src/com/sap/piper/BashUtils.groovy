package com.sap.piper

class BashUtils implements Serializable {
    static final long serialVersionUID = 1L

    static String escape(String str) {
        // put string in single quotes and escape contained single quotes by putting them into a double quoted string

        def escapedString = str.replace("'", "'\"'\"'")
        return "'${escapedString}'"
    }
}
