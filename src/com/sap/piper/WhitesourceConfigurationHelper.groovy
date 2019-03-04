package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import java.security.MessageDigest

class WhitesourceConfigurationHelper implements Serializable {

    private static def SCALA_CONTENT_KEY = "@__content"

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
            [name: 'apiKey', value: config.orgToken, warnIfPresent: true],
            [name: 'productName', value: config.productName],
            [name: 'productVersion', value: config.productVersion],
            [name: 'projectName', value: config.projectName],
            [name: 'projectVersion', value: config.productVersion],
            [name: 'productToken', value: config.productToken, omitIfPresent: 'projectToken'],
            [name: 'userKey', value: config.userKey, warnIfPresent: true],
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

    static def extendConfigurationFile(script, utils, config, path) {
        def mapping = [:]
        def parsingClosure
        def serializationClosure
        def inputFile = config.configFilePath.replaceFirst('\\./', '')
        def suffix = MessageDigest.getInstance("MD5").digest(config.configFilePath.bytes).encodeHex().toString()
        def targetFile = "${inputFile}.${suffix}"
        switch (config.scanType) {
            case 'unifiedAgent':
            case 'fileAgent':
                mapping = [
                    [name: 'apiKey', value: config.orgToken, warnIfPresent: true],
                    [name: 'productName', value: config.productName],
                    [name: 'productToken', value: config.productToken, omitIfPresent: 'projectToken'],
                    [name: 'userKey', value: config.userKey, warnIfPresent: true]
                ]
                parsingClosure = { fileReadPath -> return script.readProperties (file: fileReadPath) }
                serializationClosure = { configuration -> serializeUAConfig(configuration) }
                break
            case 'npm':
                mapping = [
                    [name: 'apiKey', value: config.orgToken, warnIfPresent: true],
                    [name: 'productName', value: config.productName],
                    [name: 'productToken', value: config.productToken, omitIfPresent: 'projectToken'],
                    [name: 'userKey', value: config.userKey, warnIfPresent: true]
                ]
                parsingClosure = { fileReadPath -> return script.readJSON (file: fileReadPath) }
                serializationClosure = { configuration -> return new JsonUtils().getPrettyJsonString(configuration) }
                break
            case 'pip':
                mapping = [
                    [name: "'org_token'", value: "\'${config.orgToken}\'", warnIfPresent: true],
                    [name: "'product_name'", value: "\'${config.productName}\'"],
                    [name: "'product_token'", value: "\'${config.productToken}\'"],
                    [name: "'user_key'", value: "\'${config.userKey}\'", warnIfPresent: true]
                ]
                parsingClosure = { fileReadPath -> return readPythonConfig (script, fileReadPath) }
                serializationClosure = { configuration -> serializePythonConfig(configuration) }
                targetFile = "${inputFile}.${suffix}.py"
                break
            case 'sbt':
                mapping = [
                    [name: "whitesourceOrgToken in ThisBuild", value: "\"${config.orgToken}\"", warnIfPresent: true],
                    [name: "whitesourceProduct in ThisBuild", value: "\"${config.productName}\""],
                    [name: "whitesourceServiceUrl in ThisBuild", value: "uri(\"${config.agentUrl}\")"]
                    // actually not supported [name: "whitesourceUserKey in ThisBuild", value: config.userKey]
                ]
                parsingClosure = { fileReadPath -> return readScalaConfig (script, mapping, fileReadPath) }
                serializationClosure = { configuration -> serializeScalaConfig (configuration) }
                targetFile = inputFile
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
                //if (entry.warnIfPresent && moduleSpecificFile[entry.name])
                    //Notify.warning(script, "Obsolete configuration ${entry.name} detected, please omit its use and rely on configuration via Piper.", 'WhitesourceConfigurationHelper')
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

    static private def readPythonConfig(script, filePath) {
        def contents = script.readFile file: filePath
        def lines = contents.split('\n')
        def resultMap = [:]
        lines.each {
            line ->
                List parts = line?.replaceAll(',$', '')?.split(':')
                def key = parts[0]?.trim()
                parts.removeAt(0)
                resultMap[key] = parts.size() > 0 ? (parts as String[]).join(':').trim() : null
        }
        return resultMap
    }

    static private def serializePythonConfig(configuration) {
        StringBuilder result = new StringBuilder()
        configuration.each {
            entry ->
                if(entry.key != '}')
                    result.append(entry.value ? '    ' : '').append(entry.key).append(entry.value ? ': ' : '').append(entry.value ?: '').append(entry.value ? ',' : '').append('\r\n')
        }
        return result.toString().replaceAll(',$', '\r\n}')
    }

    static private def readScalaConfig(script, mapping, filePath) {
        def contents = script.readFile file: filePath
        def lines = contents.split('\n')
        def resultMap = [:]
        resultMap[SCALA_CONTENT_KEY] = []
        def keys = mapping.collect( { it.name } )
        lines.each {
            line ->
                def parts = line?.split(':=').toList()
                def key = parts[0]?.trim()
                if (keys.contains(key)) {
                    resultMap[key] = parts[1]?.trim()
                } else if (line != null) {
                    resultMap[SCALA_CONTENT_KEY].add(line)
                }
        }
        return resultMap
    }

    static private def serializeScalaConfig(configuration) {
        StringBuilder result = new StringBuilder()

        // write the general content
        configuration[SCALA_CONTENT_KEY].each {
            line ->
                result.append(line)
                result.append('\r\n')
        }

        // write the mappings
        def confKeys = configuration.keySet()
        confKeys.remove(SCALA_CONTENT_KEY)

        confKeys.each {
            key ->
                def value = configuration[key]
                result.append(key)
                if (value != null) {
                    result.append(' := ').append(value)
                }
                result.append('\r\n')
        }

        return result.toString()
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
