package com.sap.piper.tools.neo

enum WarAction {

    DEPLOY('deploy'), ROLLING_UPDATE('rolling-update')

    private String value

    WarAction(String value) {
        this.value = value
    }

    static Set stringValues() {
        return values().collect { each -> each.value } as Set
    }

    static WarAction fromString(String value) {
        WarAction enumValue = values().find { each -> each.value == value }

        if (enumValue == null) {
            throw new IllegalArgumentException("${value} is not in the list of possible values ${stringValues()}")
        }

        return enumValue
    }
}
