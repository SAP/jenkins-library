package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class CredentialCollection implements Serializable {
    static final long serialVersionUID = 1L

    List credentials = []

    public CredentialCollection() {}

    @NonCPS
    def toCredentialJson() {
        return "{ \"credentials\": [\n  ${credentials.join(",\n  ")}\n]}\n"
    }

    @NonCPS
    def addCredential(Credential credential) {
        this.credentials.add(credential)
    }
}
