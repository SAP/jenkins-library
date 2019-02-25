package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class BashUtils implements Serializable {
    static final long serialVersionUID = 1L

    static String quoteAndEscape(String str) {
        // put string in single quotes and escape contained single quotes by putting them into a double quoted string

        def escapedString = str.replace("'", "'\"'\"'")
        return "'${escapedString}'"
    }

    @NonCPS
    static String escapeBlanks(CharSequence c) {
        if(! c) return c
        c.replaceAll(' ', '\\\\ ')
    }
}
