import groovy.io.FileType
import groovy.json.JsonOutput
import groovy.json.JsonSlurper
import org.yaml.snakeyaml.Yaml
import org.codehaus.groovy.control.CompilerConfiguration
import com.sap.piper.GenerateDocumentation
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.DefaultValueCache
import java.util.regex.Matcher
import groovy.text.StreamingTemplateEngine

import com.sap.piper.MapUtils

//
// Collects helper functions for rendering the documentation
//
class TemplateHelper {

    static createDependencyList(Set deps) {
        def t = ''
        t += 'The step depends on the following Jenkins plugins\n\n'
        def filteredDeps = deps.findAll { dep -> dep != 'UNIDENTIFIED' }

        if(filteredDeps.contains('kubernetes')) {
            // The docker plugin is not detected by the tests since it is not
            // handled via step call, but it is added to the environment.
            // Hovever kubernetes plugin and docker plugin are closely related,
            // hence adding docker if kubernetes is present.
            filteredDeps.add('docker')
        }

        if(filteredDeps.isEmpty()) {
            t += '* &lt;none&gt;\n'
        } else {
            filteredDeps
                .sort()
                .each { dep -> t += "* [${dep}](https://plugins.jenkins.io/${dep})\n" }
        }

        if(filteredDeps.contains('kubernetes')) {
            t += "\nThe kubernetes plugin is only used if running in a kubernetes environment."
        }

        t += '''|
                |Transitive dependencies are omitted.
                |
                |The list might be incomplete.
                |
                |Consider using the [ppiper/jenkins-master](https://cloud.docker.com/u/ppiper/repository/docker/ppiper/jenkins-master)
                |docker image. This images comes with preinstalled plugins.
                |'''.stripMargin()
        return t
    }

    static createParametersTable(Map parameters) {

        def t = ''
        t += '| name | mandatory | default | possible values |\n'
        t += '|------|-----------|---------|-----------------|\n'

        parameters.keySet().toSorted().each {

            def props = parameters.get(it)

            def defaultValue = isComplexDefault(props.defaultValue) ? renderComplexDefaultValue(props.defaultValue) : renderSimpleDefaultValue(props.defaultValue)

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

    private static renderSimpleDefaultValue(def _default) {
        if (_default == null) return ''
        return "`${_default}`"
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

    static createStageContentSection(Map stageDescriptions) {
        def t = 'This stage comprises following steps which are activated depending on your use-case/configuration:\n\n'

        t += '| step | step description |\n'
        t += '| ---- | ---------------- |\n'

        stageDescriptions.each {step, description ->
            t += "| [${step}](../steps/${step}.md) | ${description.trim()} |\n"
        }

        return t
    }

    static createStageActivationSection() {
        def t = '''This stage will be active if any one of the following conditions is met:

* Stage configuration in [config.yml file](../configuration.md) contains entries for this stage.
* Any of the conditions are met which are explained in the section [Step Activation](#step-activation).
'''
        return t.trim()
    }

    static createStepActivationSection(Map configConditions) {
        if (!configConditions) return 'For this stage no conditions are assigned to steps.'
        def t = 'Certain steps will be activated automatically depending on following conditions:\n\n'


        t += '| step | config key | config value | file pattern |\n'
        t += '| ---- | ---------- | ------------ | ------------ |\n'

        configConditions?.each {stepName, conditions ->
            t += "| ${stepName} "
            t += "| ${renderValueList(conditions?.configKeys)} "
            t += "| ${renderValueList(mapToValueList(conditions?.config))} "

            List filePatterns = []
            if (conditions?.filePattern) filePatterns.add(conditions?.filePattern)
            if (conditions?.filePatternFromConfig) filePatterns.add(conditions?.filePatternFromConfig)
            t += "| ${renderValueList(filePatterns)} |\n"
        }

        t += '''
!!! info "Step condition details"
    There are currently several conditions which can be checked.<br /> This is done in the [Init stage](init.md) of the pipeline shortly after checkout of the source code repository.<br/ >
    **Important: It will be sufficient that any one condition per step is met.**

    * `config key`: Checks if a defined configuration parameter is set.
    * `config value`: Checks if a configuration parameter has a defined value.
    * `file pattern`: Checks if files according a defined pattern exist in the project. Either the pattern is speficified direcly or it is retrieved from a configuration parameter.


!!! note "Overruling step activation conditions"
    It is possible to overrule the automatically detected step activation status.<br />
    Just add to your stage configuration `<stepName>: false`, for example `deployToKubernetes: false`.

For details about the configuration options, please see [Configuration of Piper](../configuration.md).
'''

        return t
    }

    private static renderValueList(List valueList) {
        if (!valueList) return ''
        if (valueList.size() > 1) {
            List quotedList = []
            valueList.each {listItem ->
                quotedList.add("-`${listItem}`")
            }
            return quotedList.join('<br />')
        } else {
            return "`${valueList[0]}`"
        }
    }

    private static mapToValueList(Map map) {
        List valueList = []
        map?.each {key, value ->
            if (value instanceof List) {
                value.each {listItem ->
                    valueList.add("${key}: ${listItem}")
                }
            } else {
                valueList.add("${key}: ${value}")
            }
        }
        return valueList
    }

    static createStageConfigurationSection() {
        return 'The stage parameters need to be defined in the section `stages` of [config.yml file](../configuration.md).'
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

    static Map getYamlResource(String resource) {
        def ymlContent = new File(projectRoot,"resources/${resource}").text
        return new Yaml().load(ymlContent)
    }

    static getDummyScript(def stepName) {

        def _stepName = stepName

        return  new Script() {

            def STEP_NAME = _stepName

            def handlePipelineStepErrors(def m, Closure c) {
                c()
            }

            def libraryResource(def r) {
                new File(projectRoot,"resources/${r}").text
            }

            def readYaml(def m) {
                new Yaml().load(m.text)
            }

            void echo(m) {
                println(m)
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

    static getStageStepKeys(def script) {
        try {
            return script.STAGE_STEP_KEYS ?: []
        } catch (groovy.lang.MissingPropertyException ex) {
            System.err << "[INFO] STAGE_STEP_KEYS not set for: ${script.STEP_NAME}.\n"
            return []
        }
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
                    if(method.getName() == 'call' && (method.getAnnotation(GenerateDocumentation) != null || method.getAnnotation(GenerateStageDocumentation) != null)) {
                        docuRelevantSteps << scriptName
                        break
                    }
                }
            }
        }
        docuRelevantSteps
    }

    static resolveDocuRelevantStages(GroovyScriptEngine gse, File stepsDir) {

        def docuRelevantStages = [:]

        stepsDir.traverse(type: FileType.FILES, maxDepth: 0) {
            if(it.getName().endsWith('.groovy')) {
                def scriptName = (it =~  /vars\${File.separator}(.*)\.groovy/)[0][1]
                def stepScript = gse.createScript("${scriptName}.groovy", new Binding())
                for (def method in stepScript.getClass().getMethods()) {
                    GenerateStageDocumentation stageDocsAnnotation = method.getAnnotation(GenerateStageDocumentation)
                    if(method.getName() == 'call' && stageDocsAnnotation != null) {
                        docuRelevantStages[scriptName] = stageDocsAnnotation.defaultStageName()
                        break
                    }
                }
            }
        }
        docuRelevantStages
    }
}

roots = [
    new File(Helper.projectRoot, "vars").getAbsolutePath(),
    new File(Helper.projectRoot, "src").getAbsolutePath()
]

stepsDir = null
stepsDocuDir = null
stagesDocuDir = null
customDefaults = null

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
    p longOpt: 'docuDirStages', args: 1, argName: 'dir', 'The directory containing the docu stubs for pipeline stages. Defaults to \'documentation/docs/stages\'.'
    c longOpt: 'customDefaults', args: 1, argName: 'file', 'Additional custom default configuration'
    i longOpt: 'stageInitFile', args: 1, argName: 'file', 'The file containing initialization data for step piperInitRunStageConfiguration'
    h longOpt: 'help', 'Prints this help.'
}

def options = cli.parse(args)

if(options.h) {
    System.err << "Printing help.\n"
    cli.usage()
    return
}

if(options.s){
    System.err << "[INFO] Using custom step root: ${options.s}.\n"
    stepsDir = new File(Helper.projectRoot, options.s)
}


stepsDir = stepsDir ?: new File(Helper.projectRoot, "vars")

if(options.d) {
    System.err << "[INFO] Using custom doc dir for steps: ${options.d}.\n"
    stepsDocuDir = new File(Helper.projectRoot, options.d)
}

stepsDocuDir = stepsDocuDir ?: new File(Helper.projectRoot, "documentation/docs/steps")

if(options.p) {
    System.err << "[INFO] Using custom doc dir for stages: ${options.p}.\n"
    stagesDocuDir = new File(Helper.projectRoot, options.p)
}

stagesDocuDir = stagesDocuDir ?: new File(Helper.projectRoot, "documentation/docs/stages")

if(options.c) {
    System.err << "[INFO] Using custom defaults: ${options.c}.\n"
    customDefaults = options.c
}

// retrieve default conditions for steps
Map stageConfig
if (options.i) {
    System.err << "[INFO] Using stageInitFile ${options.i}.\n"
    stageConfig = Helper.getYamlResource(options.i)
    System.err << "[INFO] Default stage configuration: ${stageConfig}.\n"
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

// find all the stages that we have to document
Map stages = Helper.resolveDocuRelevantStages(gse, stepsDir)

boolean exceptionCaught = false

def stepDescriptors = [:]
DefaultValueCache.prepare(Helper.getDummyScript('noop'),  customDefaults)
for (step in steps) {
    try {
        stepDescriptors."${step}" = handleStep(step, gse)
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

//update stepDescriptors: remove stages and put into separate stageDescriptors map
def stageDescriptors = [:]
stages.each {key, value ->
    System.err << "[INFO] Processing stage '${key}' ...\n"
    stageDescriptors."${key}" = [:] << stepDescriptors."${key}"
    stepDescriptors.remove(key)

    //add stage name to stageDescriptors
    stageDescriptors."${key}".name = value

    //add stepCondition informmation to stageDescriptors
    stageDescriptors."${key}".configConditions = stageConfig?.stages?.get(value)?.stepConditions

    //identify step keys in stages
    def stageStepKeys = Helper.getStageStepKeys(gse.createScript( "${key}.groovy", new Binding() ))

    // prepare step descriptions
    stageDescriptors."${key}".stepDescriptions = [:]
    stageDescriptors."${key}".parameters.each {paramKey, paramValue ->

        if (paramKey in stageStepKeys) {
            stageDescriptors."${key}".stepDescriptions."${paramKey}" = "${paramValue.docu ?: ''}\n"
        }
    }

    //remove details from parameter map
    stageStepKeys.each {stepKey ->
        stageDescriptors."${key}".parameters.remove(stepKey)
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

for (stage in stageDescriptors) {
    try {
        renderStage(stage.key, stage.value)
        System.err << "[INFO] Stage '${stage.key}' has been rendered.\n"
    } catch(Exception e) {
        exceptionCaught = true
        System.err << "${e.getClass().getName()} caught while rendering stage '${stage}': ${e.getMessage()}.\n"
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
        docGenConfiguration : 'Step configuration\n\n' + TemplateHelper.createStepConfigurationSection(stepProperties.parameters),
        docJenkinsPluginDependencies     : 'Dependencies\n\n' + TemplateHelper.createDependencyList(stepProperties.dependencies)
    ]

    def template = new StreamingTemplateEngine().createTemplate(theStepDocu.text)
    String text = template.make(binding)

    theStepDocu.withWriter { w -> w.write text }
}

void renderStage(stageName, stageProperties) {

    def stageFileName = stageName.indexOf('Stage') != -1 ? stageName.split('Stage')[1].toLowerCase() : stageFileName
    File theStageDocu = new File(stagesDocuDir, "${stageFileName}.md")

    if(!theStageDocu.exists()) {
        System.err << "[WARNING] stage docu input file for stage '${stageName}' is missing.\n"
        return
    }

    def binding = [
        docGenStageName     : stageProperties.name,
        docGenDescription   : stageProperties.description,
        docGenStageContent  : 'Stage Content\n\n' + TemplateHelper.createStageContentSection(stageProperties.stepDescriptions),
        docGenStageActivation: 'Stage Activation\n\n' + TemplateHelper.createStageActivationSection(),
        docGenStepActivation: 'Step Activation\n\n' + TemplateHelper.createStepActivationSection(stageProperties.configConditions),
        docGenStageParameters    : 'Additional Stage Parameters\n\n' + TemplateHelper.createParametersSection(stageProperties.parameters),
        docGenStageConfiguration : 'Configuration of Additional Stage Parameters\n\n' + TemplateHelper.createStageConfigurationSection()
    ]
    def template = new StreamingTemplateEngine().createTemplate(theStageDocu.text)
    String text = template.make(binding)

    theStageDocu.withWriter { w -> w.write text }
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

def handleStep(stepName, gse) {

    File theStep = new File(stepsDir, "${stepName}.groovy")
    File theStepDocu = new File(stepsDocuDir, "${stepName}.md")
    File theStepDeps = new File('documentation/jenkins_workspace/plugin_mapping.json')

    if (!theStepDocu.exists() && stepName.indexOf('Stage') != -1) {
        //try to get a corresponding stage documentation
        def stageName = stepName.split('Stage')[1].toLowerCase()
        theStepDocu = new File(stagesDocuDir,"${stageName}.md" )
    }

    if(!theStepDocu.exists()) {
        System.err << "[WARNING] step docu input file for step '${stepName}' is missing.\n"
        return
    }

    System.err << "[INFO] Handling step '${stepName}'.\n"

    def defaultConfig = Helper.getConfigHelper(getClass().getClassLoader(),
        roots,
        Helper.getDummyScript(stepName)).use()

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
    def step = [
        parameters:[:],
        dependencies: (Set)[],
        dependentConfig: [:]
    ]

    //
    // provide dependencies to Jenkins plugins
    if(theStepDeps.exists()) {
        def pluginDependencies = new JsonSlurper().parse(theStepDeps)
        step.dependencies.addAll(pluginDependencies[stepName].collect { k, v -> k })
    }

    //
    // START special handling for 'script' parameter
    // ... would be better if there is no special handling required ...

    step.parameters['script'] = [
        docu: 'The common script environment of the Jenkinsfile running. ' +
            'Typically the reference to the script calling the pipeline ' +
            'step is provided with the `this` parameter, as in `script: this`. ' +
            'This allows the function to access the ' +
            '`commonPipelineEnvironment` for retrieving, e.g. configuration parameters.',
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
