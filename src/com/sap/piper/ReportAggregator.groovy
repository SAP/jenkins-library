package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import groovy.text.SimpleTemplateEngine
import hudson.tasks.junit.TestResult

@Singleton
class ReportAggregator {

    String projectIdentifier = null
    String versionControlTool = null
    boolean resilienceChecked = false
    boolean onlyPublicAPIsChecked = false
    Map apiCheckExceptions = [
        nonErpDestinations : [],
        customODataServices: []
    ]

    Set qualityChecks = []
    Set securityScans = []
    Set npmAuditedAdvisories = []
    Set staticCodeScans = []
    Set testsExecutions = []
    Set performanceTestsExecutions = []

    String minimumCodeCoverage = null
    Set jacocoExcludes = []
    int lineCoverage = 0
    int branchCoverage = 0

    int failedTests = 0
    int skippedTests = 0
    int passedTests = 0


    boolean automaticVersioning = false
    boolean deploymentToNexusExecuted = false
    boolean deploymentExecuted = false

    String fileName

    def reportDeployment() {
        deploymentExecuted = true
    }

    def reportVersionControlUsed(String tool) {
        versionControlTool = tool
    }

    def reportQualityCheck(QualityCheck quality) {
        qualityChecks.add(quality.getCategory())
    }

    def reportTestExecution(QualityCheck quality) {
        reportQualityCheck(quality)
        testsExecutions.add(quality.toString())
    }

    def reportPerformanceTestExecution(QualityCheck quality) {
        reportQualityCheck(quality)
        performanceTestsExecutions.add(quality.toString())
    }

    def reportStaticCodeExecution(QualityCheck quality) {
        reportQualityCheck(quality)
        staticCodeScans.add(quality.toString())
    }

    def reportVulnerabilityScanExecution(QualityCheck quality) {
        reportQualityCheck(quality)
        securityScans.add(quality.toString())
    }

    def reportNpmSecurityScan(auditedAdvisories) {
        instance.reportVulnerabilityScanExecution(QualityCheck.NpmAudit)
        if (auditedAdvisories) {
            npmAuditedAdvisories.addAll(auditedAdvisories)
        }
    }

    def reportResilienceCheck() {
        reportQualityCheck(QualityCheck.ResilienceCheck)
        resilienceChecked = true
    }

    def reportServicesCheck(List nonErpDestinations, List customODataServices) {
        reportQualityCheck(QualityCheck.OnlyPublicAPIsCheck)
        onlyPublicAPIsChecked = true

        if (nonErpDestinations) {
            apiCheckExceptions.nonErpDestinations.addAll(nonErpDestinations)
        }

        if (customODataServices) {
            apiCheckExceptions.customODataServices.addAll(customODataServices)
        }
    }

    def reportCodeCoverageCheck(Script script, String minimumCodeCoverage, List jacocoExcludes) {
        this.minimumCodeCoverage = minimumCodeCoverage
        if (jacocoExcludes) {
            this.jacocoExcludes.addAll(jacocoExcludes)
        }

        TestResult result = script.currentBuild?.getRawBuild()?.getAction(hudson.tasks.junit.TestResultAction.class)?.getResult()

        if (result) {
            failedTests = result.getFailCount()
            skippedTests = result.getSkipCount()
            passedTests = result.getPassCount()
        }

        def coverageReport = script.currentBuild?.getRawBuild()?.getAction(hudson.plugins.jacoco.JacocoBuildAction.class)?.getResult()

        if (coverageReport) {
            lineCoverage = coverageReport.lineCoverage.percentage
            branchCoverage = coverageReport.branchCoverage.percentage
        }
    }

    def reportDeploymentToNexus() {
        deploymentToNexusExecuted = true
    }

    def reportAutomaticVersioning() {
        automaticVersioning = true
    }

    def reportProjectIdentifier(String projectIdentifier) {
        this.projectIdentifier = projectIdentifier
    }

    def generateReport(Script script) {
        String template = script.libraryResource "com.sap.piper/templates/pipeline_report.txt"

        if (!projectIdentifier) {
            script.error "This should not happen: Could not generate a certification report as the project identifier was not set"
        }

        Map binding = getProperties()
        Date now = new Date()

        binding.utcTimestamp = now.format("yyyy-MM-dd HH:mm", TimeZone.getTimeZone('UTC'))

        String fileNameTimestamp = now.format("yyyy-MM-dd-HH-mm", TimeZone.getTimeZone('UTC'))
        fileName = "pipeline_report_${fileNameTimestamp}_${projectIdentifier}.txt"

        return fillTemplate(template, binding)
    }

    @NonCPS
    private String fillTemplate(String template, binding) {
        def engine = new SimpleTemplateEngine()
        return engine.createTemplate(template).make(binding)
    }
}
