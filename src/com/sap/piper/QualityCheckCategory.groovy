package com.sap.piper

enum QualityCheckCategory {
    TestAutomation("Test Automation"),
    StaticCodeChecks("Static Code Checks"),
    SecurityScans("Security Scan"),
    PerformanceTests("Performance Tests"),

    private String label

    QualityCheckCategory(String label) {
        this.label = label
    }

    @Override
    String toString(){
        return label
    }
}
