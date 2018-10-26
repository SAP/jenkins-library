import groovy.io.FileType;
import org.yaml.snakeyaml.Yaml
import org.codehaus.groovy.control.CompilerConfiguration
import com.sap.piper.DefaultValueCache
import java.util.regex.Matcher

//
// Collects helper functions for rendering the docu
//
class TemplateHelper {

  static createParametersTable(Map parameters) {

    def t = ''
    t += '| name | mandatory | default | possible values |\n'

    parameters.keySet().toSorted().each {

      def props = parameters.get(it)
      t +=  "| `${it}` | ${props.required ? 'yes' : 'no'} | `${(props.defaultValue ?: 'n/a') }` | ${props.value ?: 'n/a'} |\n"
    }

    t
  }

  static createParameterDescriptionSection(Map parameters) {
    def t =  ''
    parameters.keySet().toSorted().each {
      def props = parameters.get(it)
      t += "* `${it}` - ${props.docu ?: 'n/a'}\n"
    }

    t
  }
}

//
// Collects generic helper functions
//
class Helper {

  static getConfigHelper(classLoader, roots) {

    def compilerConfig = new CompilerConfiguration()
        compilerConfig.setClasspathList( roots )

    new GroovyClassLoader(classLoader, compilerConfig, true)
        .parseClass(new File('src/com/sap/piper/ConfigurationHelper.groovy'))
        .newInstance()
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
      else {
        flat << (_key as String)
      }
    }
  }

  static void scanDocu(File f, Map params) {

    boolean docu = false,
            value = false,
            scanNextLineForParamName = false

    def docuLines = [],
        valueLines = []

    f.eachLine  {
      line ->

      if(scanNextLineForParamName) {
          scanNextLineForParamName = false

          Matcher m = (line =~ /.*withMandatoryProperty\(.*'(.*)'.*/)
          if(m.size() == 1 && m[0].size() == 2) {
            // otherwise there is a comment we do care for
            def param = m[0][1]
            def _docu = [], _value = []
            docuLines.each { _docu << it  }
            valueLines.each { _value << it}
            params[param].docu = _docu*.trim().join(' ')
            params[param].value = _value*.trim().join(' ')
          }

          docuLines.clear()
          valueLines.clear()
      }

      if( line.trim()  ==~ /^\/\*\*/ ) {
        docu = true
      }

      if(docu) {
        def _line = line.replaceAll('^.*\\*/?', '').trim()

        if(_line ==~ /@possibleValues.*/) {
            value = true
        }

        if(_line) {
          if(value) {
            if(_line ==~ /@possibleValues.*/)
              _line = (_line =~ /@possibleValues\s*?(.*)/)[0][1]
            valueLines << _line
          } else {
            docuLines << _line
          }
        }
      }

      if(docu && line.trim() ==~ /^\*\//) {
        docu = false
        value = false
        scanNextLineForParamName = true
      }
    }
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
outDir = null
stepsDocuDir = null

//
// assign parameters

if(args.length >= 1)
  stepsDir = new File(args[0])

stepsDir = stepsDir ?: new File('vars')

if(args.length >= 2)
  outDir = new File(args[1])

outDir = outDir ?: new File('out')

if(args.length >= 3)
  stepsDocuDir = new File(args[2])

stepsDocuDir = stepsDocuDir ?: new File('documentation/docs/steps')

// assign parameters
//

//
// sanity checks

if( ! outDir.exists() ) {
  if(! outDir.mkdirs()) {
    System.err << "Cannot create output direcrory '${outDir}'.\n"
    System.exit(1)
  }
}

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

List steps = []

//
// find all the steps we have to document
stepsDir.traverse(type: FileType.FILES, maxDepth: 0)
  { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars\/(.*)\.groovy/)[0][1] }

def gse = new GroovyScriptEngine( [ stepsDir.getName()  ] as String[] , getClass().getClassLoader() )

def prepareDefaultValuesStep = Helper.getPrepareDefaultValuesStep(gse)

boolean exceptionCaught = false

for (step in steps) {
  try {
    handleStep(step, prepareDefaultValuesStep, gse)
  } catch(Exception e) {
    exceptionCaught = true
    System.err << "${e.getClass().getName()} caught while handling step '${step}'."
  }
}
if(exceptionCaught) {
  System.err << "[ERROR] Exception caught during generating documentation. Check earlier log for details.\n"
  System.exit(1)
}


void handleStep(step, prepareDefaultValuesStep, gse) {
  File theStep = new File(stepsDir, "${step}.groovy")
  File theStepDocuInput = new File(stepsDocuDir, "${step}.md")
  File theGeneratedStepDocu = new File(outDir, "${step}.md")

  if(!theStepDocuInput.exists()) {
    System.err << "[WARNING] step docu input file for step '${step}' is missing.\n"
    return
  }

  System.err << "[INFO] Handling step '${step}'.\n"

  def defaultConfig = Helper.getConfigHelper(getClass().getClassLoader(), roots).loadStepDefaults(Helper.getDummyScript(prepareDefaultValuesStep, step)).use()

  def params = [] as Set

  //
  // scopedParameters is a map containing the scope as key and the parameters
  // defined with that scope as a set of strings.

  def scopedParameters

  try {
    scopedParameters = Helper.getScopedParameters(gse.createScript( "${step}.groovy", new Binding() ))
    scopedParameters.each { k, v -> params.addAll(v) }
  } catch(Exception e) {
    System.err << "[ERROR] Step '${step}' violates naming convention for scoped parameters: ${e}.\n"
    throw e
  }
  def requiredParameters = Helper.getRequiredParameters(theStep)

  params.addAll(requiredParameters)


  def parameters = [:]

  Helper.normalize(params).toSorted().each {

    it ->

    def parameterProperties = [
                                defaultValue: Helper.getValue(defaultConfig, it.split('/')),
                                required: requiredParameters.contains((it as String))
                              ]

    parameters.put(it, parameterProperties)

    // The scope is only defined for the first level of a hierarchical configuration.
    // If the first part is found, all nested parameters are allowed with that scope.
    def firstPart = it.split('/').head()
    scopedParameters.each { key, val ->
      parameterProperties.put(key, val.contains(firstPart))
    }
  }

  Helper.scanDocu(theStep, parameters)

  def text = theStepDocuInput.text
  text = text.replace('__PARAMETER_TABLE__', TemplateHelper.createParametersTable(parameters))
  text = text.replace('__PARAMETER_DESCRIPTION__', TemplateHelper.createParameterDescriptionSection(parameters))

  theGeneratedStepDocu.withWriter { w -> w.write text }
}

System.err << "[INFO] done.\n"
