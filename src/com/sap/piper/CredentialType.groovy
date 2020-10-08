package com.sap.piper

enum CredentialType {
    FILE('file'), TOKEN('token'), SECRET_TEXT('secretText'), USERNAME_PASSWORD('usernamePassword'), SSH('ssh')

    private final String value

    private CredentialType(String value) {
        this.value = value
    }

    @Override
    public String toString(){
        return value
    }
}
