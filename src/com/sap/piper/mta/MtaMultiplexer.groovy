package com.sap.piper.mta

class MtaMultiplexer implements Serializable {
    static Map createJobs(Script step, Map parameters, List excludeList, String jobPrefix, String buildDescriptorFile, String scanType, Closure worker) {
        Map jobs = [:]
        def filesToScan = []

        // avoid java.io.NotSerializableException: org.codehaus.groovy.util.ArrayIterator
        // see https://issues.jenkins-ci.org/browse/JENKINS-47730
        filesToScan.addAll(step.findFiles(glob: "**${File.separator}${buildDescriptorFile}")?:[])
        step.echo "Found ${filesToScan?.size()} ${scanType} descriptor files: ${filesToScan}"

        filesToScan = removeNodeModuleFiles(step, filesToScan)
        filesToScan = removeExcludedFiles(step, filesToScan, excludeList)

        for (String file : filesToScan){
            def options = [:]
            options.putAll(parameters)
            options.scanType = scanType
            options.buildDescriptorFile = file
            jobs["${jobPrefix} - ${file.replace("${File.separator}${buildDescriptorFile}",'')}"] = {worker(options)}
        }
        return jobs
    }

    static def removeNodeModuleFiles(Script step, filesToScan){
        step.echo "Excluding node modules:"
        return filesToScan.findAll({
            if(it.path.contains("node_modules${File.separator}")){
                step.echo "- Skipping ${it.path}"
                return false
            }
            return true
        })
    }

    static def removeExcludedFiles(Script step, filesToScan, List filesToExclude){
        def filteredFiles = []
        for (File file : filesToScan) {
            if(filesToExclude.contains(file.path)){
                step.echo "Skipping ${file.path}"
            }else{
                filteredFiles.add(file.path)
            }
        }
        return filteredFiles
    }
}
