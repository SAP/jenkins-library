import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.BashUtils
import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/mavenExecute.yaml'

def call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)

    validateParameter(parameters.defines, 'defines')
    validateParameter(parameters.flags, 'flags')
    validateParameter(parameters.goals, 'goals')
    validateStringParameter(parameters.pomPath)
    validateStringParameter(parameters.projectSettingsFile)
    validateStringParameter(parameters.globalSettingsFile)
    validateStringParameter(parameters.m2Path)

    List credentials = []
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)

    String output = ''
    if (parameters.returnStdout) {
        String outputFile = '.pipeline/maven_output.txt'
        if (!fileExists(outputFile)) {
            error "[$STEP_NAME] Internal error. A text file with the contents of the maven output was expected " +
                "but does not exist at '$outputFile'. " +
                "Please file a ticket at https://github.com/SAP/jenkins-library/issues/new/choose"
        }
        output = readFile(outputFile)
    }
    return output
}

private void validateParameter(parameter, name) {
    if (!parameter) {
        return
    }

    String errorMessage = "Expected parameter ${name} with value ${parameter} to be of type List, but it is ${parameter.class}. "

    // Specifically check for string-like types as the old step (v1.23.0 and before) allowed that as input
    if (parameter in CharSequence) {
        errorMessage += "This is a breaking change for mavenExecute in library version v1.24.0 which allowed strings as input for defines, flags and goals before. " +
            "To fix that, please update the interface to pass in lists, or use v1.23.0 which is the last version with the old interface. "

        if (parameter.contains(BashUtils.ESCAPED_SINGLE_QUOTE)) {
            errorMessage += "It looks like your input contains shell escaped quotes. "
        }

        error errorMessage + "Note that *no* shell escaping is allowed."
    }

    if (!parameter in List) {
        error errorMessage + "Note that *no* shell escaping is allowed."
    }

    for (int i = 0; i < parameter.size(); i++) {
        String element = parameter[i]
        validateStringParameter(element)
    }
}

private void validateStringParameter(String element) {
    if (!element) {
        return
    }

    if (element =~ /-D.*='.*'/) {
        echo "[$STEP_NAME WARNING] It looks like you passed a define in the form -Dmy.key='this is my value' in $element. Please note that the quotes might cause issues. Correct form: -Dmy.key=this is my value"
    }

    if (element.length() >= 2 && element.startsWith("'") && element.endsWith("'")) {
        echo "[$STEP_NAME WARNING] It looks like $element is quoted but it should not be quoted."
    }
}
