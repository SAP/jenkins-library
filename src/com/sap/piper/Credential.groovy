package com.sap.piper

enum Credential {
    FILE('file'), TOKEN('token'), USERNAME_PASSWORD('usernamePassword'), SSH('ssh')

    private String value

    public Credential(String value) {
        this.value = value
    }

    @Override
    public String toString(){
        return value
    }
}
