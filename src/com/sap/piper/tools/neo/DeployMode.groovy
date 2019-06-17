package com.sap.piper.tools.neo

enum DeployMode {
    MTA('mta'), WAR_PARAMS('warParams'), WAR_PROPERTIES_FILE('warPropertiesFile')

    private String value

    DeployMode(String value) {
        this.value = value
    }

    static Set stringValues() {
        return values().collect { each -> each.value } as Set
    }

    boolean isWarDeployment() {
        return this != DeployMode.MTA
    }

    static DeployMode fromString(String value) {
        DeployMode enumValue = values().find { each -> each.value == value }

        if (enumValue == null) {
            throw new IllegalArgumentException("${value} is not in the list of possible values ${stringValues()}")
        }

        return enumValue
    }
}
