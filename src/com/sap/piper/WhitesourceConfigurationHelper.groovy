package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class WhitesourceConfigurationHelper implements Serializable {

    static def extendUAConfigurationFile(script, utils, config, path) {
        def mapping = []
        def parsingClosure = { fileReadPath -> return script.readProperties (file: fileReadPath) }
        def serializationClosure = { configuration -> serializeUAConfig(configuration) }
        def inputFile = config.whitesource.configFilePath.replaceFirst('\\./', '')
        def suffix = utils.generateSha1("${path}${inputFile}")
        def targetFile = "${inputFile}.${suffix}"
        if(config.whitesource.productName.startsWith('DIST - ')) {
            mapping += [
                [name: 'checkPolicies', value: false, force: true],
                [name: 'forceCheckAllDependencies', value: false, force: true]
            ]
        } else {
            mapping += [
                [name: 'checkPolicies', value: true, force: true],
                [name: 'forceCheckAllDependencies', value: true, force: true]
            ]
        }
        if(config.verbose)
            mapping += [
                [name: 'log.level', value: 'debug'],
                [name: 'log.files.level', value: 'debug']
            ]

        mapping += [
            [name: 'apiKey', value: config.whitesource.orgToken, force: true],
            [name: 'productName', value: config.whitesource.productName, force: true],
            [name: 'productVersion', value: config.whitesource.productVersion?:'', force: true],
            [name: 'projectName', value: config.whitesource.projectName, force: true],
            [name: 'projectVersion', value: config.whitesource.productVersion?:'', force: true],
            [name: 'productToken', value: config.whitesource.productToken, omitIfPresent: 'projectToken', force: true],
            [name: 'userKey', value: config.whitesource.userKey, force: true],
            [name: 'forceUpdate', value: true, force: true],
            [name: 'offline', value: false, force: true],
            [name: 'ignoreSourceFiles', value: true, force: true],
            [name: 'resolveAllDependencies', value: false, force: true],
            [name: 'failErrorLevel', value: 'ALL', force: true],
            [name: 'case.sensitive.glob', value: false],
            [name: 'followSymbolicLinks', value: true]
        ]

        switch (config.scanType) {
            case 'pip':
                mapping += [
                    [name: 'python.resolveDependencies', value: true, force: true],
                    [name: 'python.ignoreSourceFiles', value: true, force: true],
                    [name: 'python.ignorePipInstallErrors', value: false],
                    [name: 'python.installVirtualenv', value: true],
                    [name: 'python.resolveHierarchyTree', value: true],
                    [name: 'python.requirementsFileIncludes', value: 'requirements.txt'],
                    [name: 'python.resolveSetupPyFiles', value: true],
                    [name: 'python.runPipenvPreStep', value: true],
                    [name: 'python.pipenvDevDependencies', value: true],
                    [name: 'python.IgnorePipenvInstallErrors', value: false],
                    [name: 'includes', value: '**/*.py **/*.txt'],
                    [name: 'excludes', value: '**/*sources.jar **/*javadoc.jar']
                ]
                break
            case 'sbt':
                mapping += [
                    [name: 'sbt.resolveDependencies', value: true, force: true],
                    [name: 'sbt.ignoreSourceFiles', value: true, force: true],
                    [name: 'sbt.aggregateModules', value: false, force: true],
                    [name: 'sbt.runPreStep', value: true],
                    [name: 'includes', value: '**/*.jar'],
                    [name: 'excludes', value: '**/*sources.jar **/*javadoc.jar']
                ]
                break
            case 'golang':
                mapping += [
                    [name: 'go.resolveDependencies', value: true, force: true],
                    [name: 'go.ignoreSourceFiles', value: true, force: true],
                    [name: 'go.collectDependenciesAtRuntime', value: false],
                    [name: 'go.dependencyManager', value: 'dep'],
                    [name: 'includes', value: '**/*.lock'],
                    [name: 'excludes', value: '**/*sources.jar **/*javadoc.jar']
                ]
                break
            case 'dub':
                mapping += [
                    [name: 'includes', value: '**/*.d **/*.di']
                ]
                break
            default:
                script.echo "[Warning][Whitesource] Configuration for scanType: '${config.scanType}' is not yet hardened, please do a quality assessment of your scan results."
        }

        rewriteConfiguration(script, utils, config, mapping, suffix, path, inputFile, targetFile, parsingClosure, serializationClosure)
    }

    static private def rewriteConfiguration(script, utils, config, mapping, suffix, path, inputFile, targetFile, parsingClosure, serializationClosure) {
        def inputFilePath = "${path}${inputFile}"
        def outputFilePath = "${path}${targetFile}"
        def moduleSpecificFile = parsingClosure(inputFilePath)
        if (!moduleSpecificFile && inputFilePath != config.whitesource.configFilePath)
            moduleSpecificFile = parsingClosure(config.whitesource.configFilePath)
        if (!moduleSpecificFile)
            moduleSpecificFile = [:]

        for(int i = 0; i < mapping.size(); i++) {
            def entry = mapping.get(i)
            def dependentValue = entry.omitIfPresent ? moduleSpecificFile[entry.omitIfPresent] : null
            if ((entry.omitIfPresent && !dependentValue || !entry.omitIfPresent) && (entry.force || moduleSpecificFile[entry.name] == null) && entry.value != 'null')
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
        config.whitesource.configFilePath = outputFilePath
    }

    @NonCPS
    static private def serializeUAConfig(configuration) {
        Properties p = new Properties()
        p.putAll(configuration)

        new StringWriter().with{ w -> p.store(w, null); w }.toString()
    }
}
