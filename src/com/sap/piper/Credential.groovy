package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class Credential implements Serializable {
    static final long serialVersionUID = 1L

    def alias
    def username
    def password

    public Credential(alias, username, password) {
        this.alias = alias
        this.username = username
        this.password = password
    }

    @NonCPS
    def String toString() {
        return "{\"alias\":\"" + this.alias + "\"," +
            "\"username\":\"" + this.username + "\"," +
            "\"password\":\"" + this.password + "\"}"
    }
}
