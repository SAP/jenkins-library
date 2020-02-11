package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsFileExistsRule implements TestRule {

    final BasePipelineTest testInstance
    final List existingFiles

    /**
     * The List of files that have been queried via `fileExists`
     */
    final List queriedFiles = []

    JenkinsFileExistsRule(BasePipelineTest testInstance) {
        this(testInstance,[])
    }

    JenkinsFileExistsRule(BasePipelineTest testInstance, List existingFiles) {
        this.testInstance = testInstance
        this.existingFiles = existingFiles
    }

    JenkinsFileExistsRule registerExistingFile(String file) {
        existingFiles.add(file)
        return  this
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod('fileExists', [String.class], {s ->
                    def fileName = s.toString()
                    queriedFiles.add(fileName)
                    return fileExists(fileName)
                })

                testInstance.helper.registerAllowedMethod('fileExists', [Map.class], {m ->
                    def fileName = m.file.toString()
                    queriedFiles.add(fileName)
                    return fileExists(fileName)
                })

                base.evaluate()
            }
        }
    }

    private boolean fileExists(String fileName) {
        def fileExists = fileName in existingFiles
        if (!fileExists && testInstance.binding.hasVariable('files')) {
            Map files = testInstance.binding.getVariable('files')
            fileExists = files.containsKey(fileName)
        }
        return fileExists
    }

}
