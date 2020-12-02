package com.sap.piper

enum QualityCheck {

    UnitTests("Unit Tests for the Backend", QualityCheckCategory.TestAutomation),
    BackendIntegrationTests("Backend Integration Tests", QualityCheckCategory.TestAutomation),
    FrontendIntegrationTests("Frontend Integration Tests", QualityCheckCategory.TestAutomation),
    EndToEndTests("End-To-End Tests", QualityCheckCategory.TestAutomation),
    FrontendUnitTests("Unit Tests for the Frontend", QualityCheckCategory.TestAutomation),
    GatlingTests("Performance Tests with Gatling", QualityCheckCategory.PerformanceTests),
    JMeterTests("Performance Tests with JMeter", QualityCheckCategory.PerformanceTests),
    PmdCheck("PMD Static Code Checks", QualityCheckCategory.StaticCodeChecks),
    FindbugsCheck("Findbugs Static Code Checks", QualityCheckCategory.StaticCodeChecks),
    NpmAudit("Npm Audit", QualityCheckCategory.SecurityScans),
    CheckmarxScan("Checkmarx Scan", QualityCheckCategory.SecurityScans),
    FortifyScan("Fortify Scan", QualityCheckCategory.SecurityScans),
    WhiteSourceScan("WhiteSource Scan", QualityCheckCategory.SecurityScans),
    SourceClearScan("SourceClearScan Scan", QualityCheckCategory.SecurityScans),

    private String label
    private QualityCheckCategory category

    QualityCheck(String label, QualityCheckCategory category) {
        this.label = label
        this.category = category
    }

    @Override
    String toString(){
        return label
    }

    String getCategory(){
        return category.toString()
    }
}
