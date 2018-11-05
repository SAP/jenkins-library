import groovy.io.FileType;
import org.yaml.snakeyaml.Yaml
import org.codehaus.groovy.control.CompilerConfiguration
import com.sap.piper.DefaultValueCache
import java.util.regex.Matcher

//
// Collects helper functions for rendering the docu
//
class TemplateHelper {

    static replaceParagraph(def textIn, int level, name, replacement) {

        boolean insideParagraph = false
        def textOut = ''

        textIn.eachLine {

            line ->

            if(insideParagraph && line ==~ "^#{1,${level}} .*\$") {
                insideParagraph = false
            }

            if(! insideParagraph) {
                textOut += "${line}\n"
            }

            if(line ==~ "^#{${level}} ${name}.*\$") {
                insideParagraph = true
                textOut += "${replacement}\n\n"
            }
        }

        textOut
    }

    static createParametersTable(Map parameters) {

        def t = ''
        t += '| name | mandatory | default | possible values |\n'
        t += '|------|-----------|---------|-----------------|\n'

        parameters.keySet().toSorted().each {

            def props = parameters.get(it)
            t +=  "| `${it}` | ${props.required ? 'yes' : 'no'} | ${(props.defaultValue ? '`' +  props.defaultValue + '`' : '') } | ${props.value ?: ''} |\n"
        }

        t
    }

    static createParameterDescriptionSection(Map parameters) {
        def t =  ''
        parameters.keySet().toSorted().each {
            def props = parameters.get(it)
            t += "* `${it}` - ${props.docu ?: ''}\n"
        }

        t
    }

    static createStepConfigurationSection(Map parameters) {

        def t = '''|
                   |We recommend to define values of step parameters via [config.yml file](../configuration.md).
                   |
                   |In following sections the configuration is possible:'''.stripMargin()

        t += '| parameter | general | step | stage |\n'
        t += '|-----------|---------|------|-------|\n'

        parameters.keySet().toSorted().each {
            def props = parameters.get(it)
            t += "| `${it}` | ${props.GENERAL_CONFIG ? 'X' : ''} | ${props.STEP_CONFIG ? 'X' : ''} | ${props.PARAMS ? 'X' : ''} |\n"
        }

        t
    }
}

//
// Collects generic helper functions
//
class Helper {

    static getConfigHelper(classLoader, roots, script) {

        def compilerConfig = new CompilerConfiguration()
            compilerConfig.setClasspathList( roots )

        new GroovyClassLoader(classLoader, compilerConfig, true)
            .parseClass(new File('src/com/sap/piper/ConfigurationHelper.groovy'))
            .newInstance(script, [:])
        }

    static getPrepareDefaultValuesStep(def gse) {

        def prepareDefaultValuesStep = gse.createScript('prepareDefaultValues.groovy', new Binding())

        prepareDefaultValuesStep.metaClass.handlePipelineStepErrors {
            m, c ->  c()
        }
        prepareDefaultValuesStep.metaClass.libraryResource {
            f ->  new File("resources/${f}").text
        }
        prepareDefaultValuesStep.metaClass.readYaml {
            m -> new Yaml().load(m.text)
        }

        prepareDefaultValuesStep
    }

    static getDummyScript(def prepareDefaultValuesStep, def stepName) {

        def _prepareDefaultValuesStep = prepareDefaultValuesStep
        def _stepName = stepName

        return  new Script() {

            def STEP_NAME = _stepName

            def prepareDefaultValues() {
                _prepareDefaultValuesStep()
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
                docuEnd = false

        def docuLines = [], valueLines = []

        f.eachLine  {
            line ->

            if(docuEnd) {
                docuEnd = false

                if(isHeader(line)) {
                    def _docu = []
                    docuLines.each { _docu << it  }
                    _docu = Helper.trim(_docu)
                    step.description = _docu*.trim().join('\n')
                } else {

                    def param = retrieveParameterName(line)

                    if(!param) {
                        throw new RuntimeException('Cannot retrieve parameter for a comment')
                    }

                    if(step.parameters[param].docu || step.parameters[param].value)
                        System.err << "[WARNING] There is already some documentation for parameter '${param}. Is this parameter documented twice?'\n"

                    def _docu = [], _value = []
                    docuLines.each { _docu << it  }
                    valueLines.each { _value << it}
                    step.parameters[param].docu = _docu*.trim().join(' ').trim()
                    step.parameters[param].value = _value*.trim().join(' ').trim()
                }
                docuLines.clear()
                valueLines.clear()
            }

            if( line.trim()  ==~ /^\/\*\*/ ) {
                docu = true
            }

            if(docu) {
                def _line = line
                _line = _line.replaceAll('^\\s*', '') // leading white spaces
                if(_line.startsWith('/**')) _line = _line.replaceAll('^\\/\\*\\*', '') // start comment
                if(_line.startsWith('*/')) _line = _line.replaceAll('^\\*/', '') // end comment
                if(_line.startsWith('*')) _line = _line.replaceAll('^\\*', '') // continue comment
                if(_line ==~ /.*@possibleValues.*/) {
                    value = true
                }

                if(value) {
                    if(_line) {
                        _line = (_line =~ /.*@possibleValues\s*?(.*)/)[0][1]
                        valueLines << _line
                    }
                } else {
                    docuLines << _line.trim()
                }
            }

            if(docu && line.trim() ==~ /^\*\//) {
                docu = false
                value = false
                docuEnd = true
            }
        }
    }

    private static isHeader(line) {
        Matcher headerMatcher = (line =~ /(def|void)\s*call\s*\(/ )
        return headerMatcher.size() == 1 && headerMatcher[0].size() == 2
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
        params.put('PARAMS', script.PARAMETER_KEYS ?: [] )

        return params
    }

    static getRequiredParameters(File f) {
        def params = [] as Set
        f.eachLine  {
            line ->
            if( line ==~ /.*withMandatoryProperty.*/ ) {
                def param = (line =~ /.*withMandatoryProperty\('(.*)'/)[0][1]
                params << param
            }
        }
        return params
    }

    static getValue(Map config, def pPath) {
        def p =config[pPath.head()]
        if(pPath.size() == 1) return p // there is no tail
        if(p in Map) getValue(p, pPath.tail())
        else return p
    }
}

roots = [
    'vars',
    'src',
    ]

stepsDir = null
stepsDocuDir = null

steps = []

//
// assign parameters

if(args.length >= 1)
    stepsDir = new File(args[0])

stepsDir = stepsDir ?: new File('vars')

if(args.length >= 2)
    stepsDocuDir = new File(args[1])

stepsDocuDir = stepsDocuDir ?: new File('documentation/docs/steps')


if(args.length >= 3)
    steps << args[2]

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


//
// find all the steps we have to document (if no step has been provided from outside)
if( ! steps) {
    stepsDir.traverse(type: FileType.FILES, maxDepth: 0)
        { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars\/(.*)\.groovy/)[0][1] }
} else {
    System.err << "[INFO] Generating docu only for step ${steps.size > 1 ? 's' : ''} ${steps}.\n"
}

def gse = new GroovyScriptEngine( [ stepsDir.getName()  ] as String[] , getClass().getClassLoader() )

def prepareDefaultValuesStep = Helper.getPrepareDefaultValuesStep(gse)

boolean exceptionCaught = false

def stepDescriptors = [:]
for (step in steps) {
    try {
        stepDescriptors."${step}" = handleStep(step, prepareDefaultValuesStep, gse)
    } catch(Exception e) {
        exceptionCaught = true
        System.err << "${e.getClass().getName()} caught while handling step '${step}': ${e.getMessage()}.\n"
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

System.err << "[INFO] done.\n"

void renderStep(stepName, stepProperties) {

    File theStepDocu = new File(stepsDocuDir, "${stepName}.md")

    if(!theStepDocu.exists()) {
        System.err << "[WARNING] step docu input file for step '${stepName}' is missing.\n"
        return
    }

    def text = theStepDocu.text
    if(stepProperties.description) {
        text = TemplateHelper.replaceParagraph(text, 2, 'Description', '\n' + stepProperties.description)
    }
    if(stepProperties.parameters) {

        text = TemplateHelper.replaceParagraph(text, 2, 'Parameters', '\n' +
                TemplateHelper.createParametersTable(stepProperties.parameters) + '\n' +
                TemplateHelper.createParameterDescriptionSection(stepProperties.parameters))


        text = TemplateHelper.replaceParagraph(text, 2, 'Step configuration', '\n' +
                TemplateHelper.createStepConfigurationSection(stepProperties.parameters))
    }
    theStepDocu.withWriter { w -> w.write text }
}

def handleStep(stepName, prepareDefaultValuesStep, gse) {

    File theStep = new File(stepsDir, "${stepName}.groovy")
    File theStepDocu = new File(stepsDocuDir, "${stepName}.md")

    if(!theStepDocu.exists()) {
        System.err << "[WARNING] step docu input file for step '${stepName}' is missing.\n"
        return
    }

    System.err << "[INFO] Handling step '${stepName}'.\n"

    def defaultConfig = Helper.getConfigHelper(getClass().getClassLoader(),
                                               roots,
                                               Helper.getDummyScript(prepareDefaultValuesStep, stepName)).use()

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

    def step = [parameters:[:]]

    //
    // START special handling for 'script' parameter
    // ... would be better if there is no special handling required ...

    step.parameters['script'] = [
                               docu: 'The common script environment of the Jenkinsfile running. ' +
                                     'Typically the reference to the script calling the pipeline ' +
                                     'step is provided with the this parameter, as in script: this. ' +
                                     'This allows the function to access the ' +
                                     'commonPipelineEnvironment for retrieving, for example, configuration parameters.',
                               required: true,

                               GENERAL_CONFIG: 'false',
                               STEP_CONFIG: 'false',
                               PARAMS: 'true'
                             ]

    // END special handling for 'script' parameter

    Helper.normalize(params).toSorted().each {

        it ->

            def parameterProperties = [
                                        defaultValue: Helper.getValue(defaultConfig, it.split('/')),
                                        required: requiredParameters.contains((it as String))
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

    step
}
