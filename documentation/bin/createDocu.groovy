import groovy.io.FileType
import groovy.json.JsonOutput
import org.yaml.snakeyaml.Yaml
import org.codehaus.groovy.control.CompilerConfiguration
import com.sap.piper.GenerateDocumentation
import java.util.regex.Matcher
import groovy.text.StreamingTemplateEngine

import com.sap.piper.MapUtils

//
// Collects helper functions for rendering the documentation
//
class TemplateHelper {

    static createParametersTable(Map parameters) {

        def t = ''
        t += '| name | mandatory | default | possible values |\n'
        t += '|------|-----------|---------|-----------------|\n'

        parameters.keySet().toSorted().each {

            def props = parameters.get(it)

            def defaultValue = isComplexDefault(props.defaultValue) ? renderComplexDefaultValue(props.defaultValue) :
                                props.defaultValue != null ? "`${props.defaultValue}`" : ''

            t +=  "| `${it}` | ${props.mandatory ?: props.required ? 'yes' : 'no'} | ${defaultValue} | ${props.value ?: ''} |\n"
        }

        t
    }

    private static boolean isComplexDefault(def _default) {
        if(! (_default in Collection)) return false
        if(_default.size() == 0) return false
        for(def entry in _default) {
            if(! (entry in Map)) return false
            if(! entry.dependentParameterKey) return false
            if(! entry.key) return false
        }
        return true
    }

    private static renderComplexDefaultValue(def _default) {
        _default
            .collect { "${it.dependentParameterKey}=`${it.key ?: '<empty>'}`:`${it.value ?: '<empty>'}`" }
            .join('<br />')
    }

    static createParameterDescriptionSection(Map parameters) {
        def t =  ''
        parameters.keySet().toSorted().each {
            def props = parameters.get(it)
            t += "* `${it}` - ${props.docu ?: ''}\n"
        }

        t.trim()
    }

    static createParametersSection(Map parameters) {
        createParametersTable(parameters) + '\n' + createParameterDescriptionSection(parameters)
    }

    static createStepConfigurationSection(Map parameters) {

        def t = '''|We recommend to define values of step parameters via [config.yml file](../configuration.md).
                   |
                   |In following sections of the config.yml the configuration is possible:\n\n'''.stripMargin()

        t += '| parameter | general | step/stage |\n'
        t += '|-----------|---------|------------|\n'

        parameters.keySet().toSorted().each {
            def props = parameters.get(it)
            t += "| `${it}` | ${props.GENERAL_CONFIG ? 'X' : ''} | ${props.STEP_CONFIG ? 'X' : ''} |\n"
        }

        t.trim()
    }
}

//
// Collects generic helper functions
//
class Helper {

    static projectRoot = new File(Helper.class.protectionDomain.codeSource.location.path).getParentFile().getParentFile().getParentFile()

    static getConfigHelper(classLoader, roots, script) {

        def compilerConfig = new CompilerConfiguration()
        compilerConfig.setClasspathList( roots )

        new GroovyClassLoader(classLoader, compilerConfig, true)
            .parseClass(new File(projectRoot, 'src/com/sap/piper/ConfigurationHelper.groovy'))
            .newInstance(script, [:]).loadStepDefaults()
    }

    static getPrepareDefaultValuesStep(def gse) {

        def prepareDefaultValuesStep = gse.createScript('prepareDefaultValues.groovy', new Binding())

        prepareDefaultValuesStep.metaClass.handlePipelineStepErrors {
            m, c ->  c()
        }
        prepareDefaultValuesStep.metaClass.libraryResource {
            f ->  new File(projectRoot,"resources/${f}").text
        }
        prepareDefaultValuesStep.metaClass.readYaml {
            m -> new Yaml().load(m.text)
        }
        prepareDefaultValuesStep.metaClass.echo {
            m -> println(m)
        }


        prepareDefaultValuesStep
    }

    static getDummyScript(def prepareDefaultValuesStep, def stepName, Map prepareDefaultValuesStepParams) {

        def _prepareDefaultValuesStep = prepareDefaultValuesStep
        def _stepName = stepName

        return  new Script() {

            def STEP_NAME = _stepName

            def prepareDefaultValues() {
                _prepareDefaultValuesStep(prepareDefaultValuesStepParams)

            }

            def run() {
                throw new UnsupportedOperationException()
            }
        }
    }

    static trim(List lines) {

        removeLeadingEmptyLines(
            removeLeadingEmptyLines(lines.reverse())
                .reverse())
    }

    private static removeLeadingEmptyLines(lines) {

        def _lines = new ArrayList(lines), trimmed = []

        boolean empty = true

        _lines.each() {

            if(empty &&  ! it.trim()) return
            empty = false
            trimmed << it
        }

        trimmed
    }

    private static normalize(Set p) {

        def normalized = [] as Set

        def interim = [:]
        p.each {
            def parts = it.split('/') as List
            _normalize(parts, interim)
        }

        interim.each { k, v -> flatten (normalized, k, v)   }

        normalized
    }

    private static void _normalize(List parts, Map interim) {
        if( parts.size >= 1) {
            if( ! interim[parts.head()]) interim[parts.head()] = [:]
            _normalize(parts.tail(), interim[parts.head()])
        }
    }

    private static flatten(Set flat, def key, Map interim) {

        if( ! interim ) flat << (key as String)

        interim.each { k, v ->

            def _key = "${key}/${k}"

            if( v && v.size() > 0 )
                flatten(flat, _key, v)
            else
                flat << (_key as String)

        }
    }

    static void scanDocu(File f, Map step) {

        boolean docu = false,
                value = false,
                mandatory = false,
                parentObject = false,
                docuEnd = false

        def docuLines = [], valueLines = [], mandatoryLines = [], parentObjectLines = []

        f.eachLine  {
            line ->

                if(line ==~ /.*dependingOn.*/) {
                    def dependentConfigKey = (line =~ /.*dependingOn\('(.*)'\).mixin\('(.*)'/)[0][1]
                    def configKey = (line =~ /.*dependingOn\('(.*)'\).mixin\('(.*)'/)[0][2]
                    if(! step.dependentConfig[configKey]) {
                        step.dependentConfig[configKey] = []
                    }
                    step.dependentConfig[configKey] << dependentConfigKey
                }

                if(docuEnd) {
                    docuEnd = false

                    if(isHeader(line)) {
                        def _docu = []
                        docuLines.each { _docu << it  }
                        _docu = Helper.trim(_docu)
                        step.description = _docu.join('\n')
                    } else {

                        def param = retrieveParameterName(line)

                        if(!param) {
                            throw new RuntimeException('Cannot retrieve parameter for a comment')
                        }

                        def _docu = [], _value = [], _mandatory = [], _parentObject = []
                        docuLines.each { _docu << it  }
                        valueLines.each { _value << it }
                        mandatoryLines.each { _mandatory << it }
                        parentObjectLines.each { _parentObject << it }
                        _parentObject << param
                        param = _parentObject*.trim().join('/').trim()

                        if(step.parameters[param].docu || step.parameters[param].value)
                            System.err << "[WARNING] There is already some documentation for parameter '${param}. Is this parameter documented twice?'\n"

                        step.parameters[param].docu = _docu*.trim().join(' ').trim()
                        step.parameters[param].value = _value*.trim().join(' ').trim()
                        step.parameters[param].mandatory = _mandatory*.trim().join(' ').trim()
                    }
                    docuLines.clear()
                    valueLines.clear()
                    mandatoryLines.clear()
                    parentObjectLines.clear()
                }

                if( line.trim()  ==~ /^\/\*\*.*/ ) {
                    docu = true
                }

                if(docu) {
                    def _line = line
                    _line = _line.replaceAll('^\\s*', '') // leading white spaces
                    if(_line.startsWith('/**')) _line = _line.replaceAll('^\\/\\*\\*', '') // start comment
                    if(_line.startsWith('*/') || _line.trim().endsWith('*/')) _line = _line.replaceAll('^\\*/', '').replaceAll('\\*/\\s*$', '') // end comment
                    if(_line.startsWith('*')) _line = _line.replaceAll('^\\*', '') // continue comment
                    if(_line.startsWith(' ')) _line = _line.replaceAll('^\\s', '')
                    if(_line ==~ /.*@possibleValues.*/) {
                        mandatory = false // should be something like reset attributes
                        value = true
                        parentObject = false
                    }
                    // some remark for mandatory e.g. some parameters are only mandatory under certain conditions
                    if(_line ==~ /.*@mandatory.*/) {
                        value = false // should be something like reset attributes ...
                        mandatory = true
                        parentObject = false
                    }
                    // grouping config properties within a parent object for easier readability
                    if(_line ==~ /.*@parentConfigKey.*/) {
                        value = false // should be something like reset attributes ...
                        mandatory = false
                        parentObject = true
                    }

                    if(value) {
                        if(_line) {
                            _line = (_line =~ /.*@possibleValues\s*?(.*)/)[0][1]
                            valueLines << _line
                        }
                    }

                    if(mandatory) {
                        if(_line) {
                            _line = (_line =~ /.*@mandatory\s*?(.*)/)[0][1]
                            mandatoryLines << _line
                        }
                    }

                    if(parentObject) {
                        if(_line) {
                            _line = (_line =~ /.*@parentConfigKey\s*?(.*)/)[0][1]
                            parentObjectLines << _line
                        }
                    }

                    if(!value && !mandatory && !parentObject) {
                        docuLines << _line
                    }
                }

                if(docu && line.trim() ==~ /^.*\*\//) {
                    docu = false
                    value = false
                    mandatory = false
                    parentObject = false
                    docuEnd = true
                }
        }
    }

    private static isHeader(line) {
        Matcher headerMatcher = (line =~ /(?:(?:def|void)\s*call\s*\()|(?:@.*)/ )
        return headerMatcher.size() == 1
    }

    private static retrieveParameterName(line) {
        Matcher m = (line =~ /.*'(.*)'.*/)
        if(m.size() == 1 && m[0].size() == 2)
            return m[0][1]
        return null
    }

    static getScopedParameters(def script) {

        def params = [:]

        params.put('STEP_CONFIG', script.STEP_CONFIG_KEYS ?: [])
        params.put('GENERAL_CONFIG', script.GENERAL_CONFIG_KEYS ?: [] )
        params.put('STAGE_CONFIG', script.PARAMETER_KEYS ?: [] )

        return params
    }

    static getRequiredParameters(File f) {
        def params = [] as Set
        f.eachLine  {
            line ->
                if (line ==~ /.*withMandatoryProperty.*/) {
                    def param = (line =~ /.*withMandatoryProperty\('(.*)'/)[0][1]
                    params << param
                }
        }
        return params
    }

    static getParentObjectMappings(File f) {
        def mappings = [:]
        def parentObjectKey = ''
        f.eachLine  {
            line ->
                if (line ==~ /.*parentConfigKey.*/ && !parentObjectKey) {
                    def param = (line =~ /.*parentConfigKey\s*?(.*)/)[0][1]
                    parentObjectKey = param.trim()
                } else if (line ==~ /\s*?(.*)[,]{0,1}/ && parentObjectKey) {
                    def pName = retrieveParameterName(line)
                    if(pName) {
                        mappings.put(pName, parentObjectKey)
                        parentObjectKey = ''
                    }
                }
        }
        return mappings
    }

    static resolveDocuRelevantSteps(GroovyScriptEngine gse, File stepsDir) {

        def docuRelevantSteps = []

        stepsDir.traverse(type: FileType.FILES, maxDepth: 0) {
            if(it.getName().endsWith('.groovy')) {
                def scriptName = (it =~  /vars\${File.separator}(.*)\.groovy/)[0][1]
                def stepScript = gse.createScript("${scriptName}.groovy", new Binding())
                for (def method in stepScript.getClass().getMethods()) {
                    if(method.getName() == 'call' && method.getAnnotation(GenerateDocumentation) != null) {
                        docuRelevantSteps << scriptName
                        break
                    }
                }
            }
        }
        docuRelevantSteps
    }
}

roots = [
    new File(Helper.projectRoot, "vars").getAbsolutePath(),
    new File(Helper.projectRoot, "src").getAbsolutePath()
]

stepsDir = null
stepsDocuDir = null
String customDefaults = null

steps = []

//
// assign parameters


def cli = new CliBuilder(
    usage: 'groovy createDocu [<options>]',
    header: 'Options:',
    footer: 'Copyright: SAP SE')

cli.with {
    s longOpt: 'stepsDir', args: 1, argName: 'dir', 'The directory containing the steps. Defaults to \'vars\'.'
    d longOpt: 'docuDir', args: 1, argName: 'dir', 'The directory containing the docu stubs. Defaults to \'documentation/docs/steps\'.'
    c longOpt: 'customDefaults', args: 1, argName: 'file', 'Additional custom default configuration'
    h longOpt: 'help', 'Prints this help.'
}

def options = cli.parse(args)

if(options.h) {
    System.err << "Printing help.\n"
    cli.usage()
    return
}

if(options.s)
    stepsDir = new File(Helper.projectRoot, options.s)

stepsDir = stepsDir ?: new File(Helper.projectRoot, "vars")

if(options.d)
    stepsDocuDir = new File(Helper.projectRoot, options.d)

stepsDocuDir = stepsDocuDir ?: new File(Helper.projectRoot, "documentation/docs/steps")

if(options.c) {
    customDefaults = options.c
}

steps.addAll(options.arguments())

// assign parameters
//

//
// sanity checks

if( !stepsDocuDir.exists() ) {
    System.err << "Steps docu dir '${stepsDocuDir}' does not exist.\n"
    System.exit(1)
}

if( !stepsDir.exists() ) {
    System.err << "Steps dir '${stepsDir}' does not exist.\n"
    System.exit(1)
}

// sanity checks
//

def gse = new GroovyScriptEngine([ stepsDir.getAbsolutePath()  ] as String[], GenerateDocumentation.class.getClassLoader() )

//
// find all the steps we have to document (if no step has been provided from outside)
if( ! steps) {
    steps = Helper.resolveDocuRelevantSteps(gse, stepsDir)
} else {
    System.err << "[INFO] Generating docu only for step ${steps.size > 1 ? 's' : ''} ${steps}.\n"
}

def prepareDefaultValuesStep = Helper.getPrepareDefaultValuesStep(gse)

boolean exceptionCaught = false

def stepDescriptors = [:]
for (step in steps) {
    try {
        stepDescriptors."${step}" = handleStep(step, prepareDefaultValuesStep, gse, customDefaults)
    } catch(Exception e) {
        exceptionCaught = true
        System.err << "${e.getClass().getName()} caught while handling step '${step}': ${e.getMessage()}.\n"
    }
}

// replace @see tag in docu by docu from referenced step.
for(step in stepDescriptors) {
    if(step.value.parameters) {
        for(param in step.value.parameters) {
            if( param?.value?.docu?.contains('@see')) {
                def otherStep = param.value.docu.replaceAll('@see', '').trim()
                param.value.docu = fetchTextFrom(otherStep, param.key, stepDescriptors)
                param.value.mandatory = fetchMandatoryFrom(otherStep, param.key, stepDescriptors)
                if(! param.value.value)
                    param.value.value = fetchPossibleValuesFrom(otherStep, param.key, stepDescriptors)
            }
        }
    }
}

for(step in stepDescriptors) {
    try {
        renderStep(step.key, step.value)
        System.err << "[INFO] Step '${step.key}' has been rendered.\n"
    } catch(Exception e) {
        exceptionCaught = true
        System.err << "${e.getClass().getName()} caught while rendering step '${step}': ${e.getMessage()}.\n"
    }
}

if(exceptionCaught) {
    System.err << "[ERROR] Exception caught during generating documentation. Check earlier log for details.\n"
    System.exit(1)
}

File docuMetaData = new File('target/docuMetaData.json')
if(docuMetaData.exists()) docuMetaData.delete()
docuMetaData << new JsonOutput().toJson(stepDescriptors)

System.err << "[INFO] done.\n"

void renderStep(stepName, stepProperties) {

    File theStepDocu = new File(stepsDocuDir, "${stepName}.md")

    if(!theStepDocu.exists()) {
        System.err << "[WARNING] step docu input file for step '${stepName}' is missing.\n"
        return
    }

    def binding = [
        docGenStepName      : stepName,
        docGenDescription   : 'Description\n\n' + stepProperties.description,
        docGenParameters    : 'Parameters\n\n' + TemplateHelper.createParametersSection(stepProperties.parameters),
        docGenConfiguration : 'Step configuration\n\n' + TemplateHelper.createStepConfigurationSection(stepProperties.parameters)
    ]
    def template = new StreamingTemplateEngine().createTemplate(theStepDocu.text)
    String text = template.make(binding)

    theStepDocu.withWriter { w -> w.write text }
}

def fetchTextFrom(def step, def parameterName, def steps) {
    try {
        def docuFromOtherStep = steps[step]?.parameters[parameterName]?.docu
        if(! docuFromOtherStep) throw new IllegalStateException("No docu found for parameter '${parameterName}' in step ${step}.")
        return docuFromOtherStep
    } catch(e) {
        System.err << "[ERROR] Cannot retrieve docu for parameter ${parameterName} from step ${step}.\n"
        throw e
    }
}

def fetchMandatoryFrom(def step, def parameterName, def steps) {
    try {
        return steps[step]?.parameters[parameterName]?.mandatory
    } catch(e) {
        System.err << "[ERROR] Cannot retrieve docu for parameter ${parameterName} from step ${step}.\n"
        throw e
    }
}

def fetchPossibleValuesFrom(def step, def parameterName, def steps) {
    return steps[step]?.parameters[parameterName]?.value ?: ''
}

def handleStep(stepName, prepareDefaultValuesStep, gse, customDefaults) {

    File theStep = new File(stepsDir, "${stepName}.groovy")
    File theStepDocu = new File(stepsDocuDir, "${stepName}.md")

    if(!theStepDocu.exists()) {
        System.err << "[WARNING] step docu input file for step '${stepName}' is missing.\n"
        return
    }

    System.err << "[INFO] Handling step '${stepName}'.\n"

    Map prepareDefaultValuesStepParams = [:]
    if (customDefaults)
        prepareDefaultValuesStepParams.customDefaults = customDefaults

    def defaultConfig = Helper.getConfigHelper(getClass().getClassLoader(),
        roots,
        Helper.getDummyScript(prepareDefaultValuesStep, stepName, prepareDefaultValuesStepParams)).use()

    def params = [] as Set

    //
    // scopedParameters is a map containing the scope as key and the parameters
    // defined with that scope as a set of strings.

    def scopedParameters

    try {
        scopedParameters = Helper.getScopedParameters(gse.createScript( "${stepName}.groovy", new Binding() ))
        scopedParameters.each { k, v -> params.addAll(v) }
    } catch(Exception e) {
        System.err << "[ERROR] Step '${stepName}' violates naming convention for scoped parameters: ${e}.\n"
        throw e
    }
    def requiredParameters = Helper.getRequiredParameters(theStep)

    params.addAll(requiredParameters)

    // translate parameter names according to compatibility annotations
    def parentObjectMappings = Helper.getParentObjectMappings(theStep)
    def compatibleParams = [] as Set
    if(parentObjectMappings) {
        params.each {
            if (parentObjectMappings[it])
                compatibleParams.add(parentObjectMappings[it] + '/' + it)
            else
                compatibleParams.add(it)
        }
        if (compatibleParams)
            params = compatibleParams
    }

    // 'dependentConfig' is only present here for internal reasons and that entry is removed at
    // end of method.
    def step = [parameters:[:], dependentConfig: [:]]

    //
    // START special handling for 'script' parameter
    // ... would be better if there is no special handling required ...

    step.parameters['script'] = [
        docu: 'The common script environment of the Jenkinsfile running. ' +
            'Typically the reference to the script calling the pipeline ' +
            'step is provided with the this parameter, as in `script: this`. ' +
            'This allows the function to access the ' +
            'commonPipelineEnvironment for retrieving, for example, configuration parameters.',
        required: true,

        GENERAL_CONFIG: false,
        STEP_CONFIG: false
    ]

    // END special handling for 'script' parameter

    Helper.normalize(params).toSorted().each {

        it ->

            def defaultValue = MapUtils.getByPath(defaultConfig, it)

            def parameterProperties =   [
                defaultValue: defaultValue,
                required: requiredParameters.contains((it as String)) && defaultValue == null
            ]

            step.parameters.put(it, parameterProperties)

            // The scope is only defined for the first level of a hierarchical configuration.
            // If the first part is found, all nested parameters are allowed with that scope.
            def firstPart = it.split('/').head()
            scopedParameters.each { key, val ->
                parameterProperties.put(key, val.contains(firstPart))
            }
    }

    Helper.scanDocu(theStep, step)

    step.parameters.each { k, v ->
        if(step.dependentConfig.get(k)) {

            def dependentParameterKey = step.dependentConfig.get(k)[0]
            def dependentValues = step.parameters.get(dependentParameterKey)?.value

            if (dependentValues) {
                def the_defaults = []
                dependentValues
                    .replaceAll('[\'"` ]', '')
                    .split(',').each {possibleValue ->
                    if (!possibleValue instanceof Boolean && defaultConfig.get(possibleValue)) {
                        the_defaults <<
                            [
                                dependentParameterKey: dependentParameterKey,
                                key: possibleValue,
                                value: MapUtils.getByPath(defaultConfig.get(possibleValue), k)
                            ]
                    }
                }
                v.defaultValue = the_defaults
            }
        }
    }

    //
    // 'dependentConfig' is only present for internal purposes and must not be used outside.
    step.remove('dependentConfig')

    step
}
