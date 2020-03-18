package com.sap.piper

enum QualityCheckCategory {
    TestAutomation("Test Automation"),
    StaticCodeChecks("Static Code Checks"),
    SecurityScans("Security Scan"),
    PerformanceTests("Performance Tests"),
    S4sdkQualityChecks("SAP Cloud SDK Quality Checks")

    private String label

    QualityCheckCategory(String label) {
        this.label = label
    }

    @Override
    String toString(){
        return label
    }
}
