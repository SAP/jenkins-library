package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import java.security.MessageDigest

class WhitesourceConfigurationHelper implements Serializable {

    static def extendUAConfigurationFile(script, utils, config, path) {
        def mapping = []
        def parsingClosure = { fileReadPath -> return script.readProperties (file: fileReadPath) }
        def serializationClosure = { configuration -> serializeUAConfig(configuration) }
        def inputFile = config.configFilePath.replaceFirst('\\./', '')
        def suffix = MessageDigest.getInstance("MD5").digest(config.configFilePath.bytes).encodeHex().toString()
        def targetFile = "${inputFile}.${suffix}"
        if(config.productName.startsWith('DIST - ')) {
            mapping += [
                [name: 'checkPolicies', value: false],
                [name: 'forceCheckAllDependencies', value: false]
            ]
        } else if(config.productName.startsWith('SHC - ')) {
            mapping += [
                [name: 'checkPolicies', value: true],
                [name: 'forceCheckAllDependencies', value: true]
            ]
        }
        if(config.verbose)
            mapping += [name: 'log.level', value: 'debug']

        mapping += [
            [name: 'apiKey', value: config.orgToken],
            [name: 'productName', value: config.productName],
            [name: 'productVersion', value: config.productVersion],
            [name: 'projectName', value: config.projectName],
            [name: 'projectVersion', value: config.productVersion],
            [name: 'productToken', value: config.productToken, omitIfPresent: 'projectToken'],
            [name: 'userKey', value: config.userKey],
            [name: 'forceUpdate', value: true],
            [name: 'offline', value: false],
            [name: 'ignoreSourceFiles', value: true],
            [name: 'resolveAllDependencies', value: false]
        ]
        switch (config.scanType) {
            case 'maven':
                mapping += [

                ]
                break
            case 'npm':
                mapping += [

                ]
                break
            case 'pip':
                mapping += [
                    [name: 'python.resolveDependencies', value: true],
                    [name: 'python.ignoreSourceFiles', value: true],
                    [name: 'python.ignorePipInstallErrors', value: false],
                    [name: 'python.installVirtualenv', value: true],
                    [name: 'python.resolveHierarchyTree', value: true],
                    [name: 'python.requirementsFileIncludes', value: 'requirements.txt'],
                    [name: 'python.resolveSetupPyFiles', value: true],
                    [name: 'python.runPipenvPreStep', value: true],
                    [name: 'python.pipenvDevDependencies', value: true],
                    [name: 'python.IgnorePipenvInstallErrors', value: false],
                    [name: 'includes', value: '**/*.py **/*.txt'],
                    [name: 'excludes', value: '**/*sources.jar **/*javadoc.jar'],
                    [name: 'case.sensitive.glob', value: false],
                    [name: 'followSymbolicLinks', value: true]
                ]
                break
            case 'sbt':
                mapping += [

                ]
                break
        }

        rewriteConfiguration(script, utils, config, mapping, suffix, path, inputFile, targetFile, parsingClosure, serializationClosure)
    }

    static private def rewriteConfiguration(script, utils, config, mapping, suffix, path, inputFile, targetFile, parsingClosure, serializationClosure) {
        def inputFilePath = "${path}${inputFile}"
        def outputFilePath = "${path}${targetFile}"
        def moduleSpecificFile = parsingClosure(inputFilePath)
        if (!moduleSpecificFile)
            moduleSpecificFile = parsingClosure(config.configFilePath)
        if (!moduleSpecificFile)
            moduleSpecificFile = [:]

        mapping.each {
            entry ->
                def dependentValue = entry.omitIfPresent ? moduleSpecificFile[entry.omitIfPresent] : null
                if ((entry.omitIfPresent && !dependentValue || !entry.omitIfPresent) && entry.value && entry.value != 'null' && entry.value != '')
                    moduleSpecificFile[entry.name] = entry.value.toString()
        }

        def output = serializationClosure(moduleSpecificFile)

        if(config.verbose)
            script.echo "Writing config file ${outputFilePath} with content:\n${output}"
        script.writeFile file: outputFilePath, text: output
        if(config.stashContent && config.stashContent.size() > 0) {
            def stashName = "modified whitesource config ${suffix}".toString()
            utils.stashWithMessage (
                stashName,
                "Stashing modified Whitesource configuration",
                outputFilePath.replaceFirst('\\./', '')
            )
            config.stashContent += [stashName]
        }
        config.configFilePath = outputFilePath
    }

    @NonCPS
    static private def serializeUAConfig(configuration) {
        Properties p = new Properties()
        configuration.each {
            entry ->
                p.setProperty(entry.key, entry.value)
        }

        new StringWriter().with{ w -> p.store(w, null); w }.toString()
    }
}
