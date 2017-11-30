package com.sap.piper

class BashUtils implements Serializable {
    static final long serialVersionUID = 1L

    static String escape(String str) {
        return "'${str.replace("\'", "\\\'")}'"
    }
}
