import groovy.json.JsonSlurper

def cli = new CliBuilder(
    usage: 'groovy createDocu [<options>]',
    header: 'Options:',
    footer: 'Copyright: SAP SE')

cli.with {
    i longOpt: 'in', args: 1, argName: 'file', 'The file containing the mapping as created by the unit tests..'
    o longOpt: 'out', args: 1, argName: 'file', 'The file containing the condenced mappings.'
    h longOpt: 'help', 'Prints this help.'
}

def options = cli.parse(args)

if(options.h) {
    System.err << "Printing help.\n"
    cli.usage()
    return
}

if(! options.i) {
    System.err << "No input file"
    cli.usage()
    return
}
if(! options.o) {
    System.err << "No output file"
    cli.usage()
    return
}

def steps = new JsonSlurper().parseText(new File(options.i).text)

def piperSteps = steps.piperSteps
def calls = steps.calls

// only temporary in order to avoid manipulating the map during
// iterating over it.
def tmpCalls = [:]

// Adjust naming
calls.each { c ->
    tmpCalls.put(retrieveStepName(c.key), c.value as Set)
}

calls = tmpCalls
tmpCalls = null

// Remove selfs
calls.each { c ->
    c.value.remove(c.key)
}

int counter=0

def alreadyHandled = []

//
// in case we exceed the value we assume some cyclic call
// between plugin steps.
int MAX_LOOP = 1600

boolean done = false

while(counter < MAX_LOOP) {

    def hereWeNeedToReplace = null
    def toBeReplaced = null

    if(alreadyHandled.size() == calls.size()) {
        done = true
        break
    }

    for (def call in calls.entrySet()) {

        stepName = call.key
        calledSteps = call.value

        if(alreadyHandled.contains(stepName)) {
            continue
        }

        for (def calledStep in calledSteps) {

            if(! ( calledStep in Map)) {

                // in case the calledStep is a map the map
                // was introduced in an earlier loop.
                // This means this entry is already handled.

                if(calledStep in piperSteps) {
                    toBeReplaced = calledStep
                    hereWeNeedToReplace = calledSteps
                    break
                }
            }
        }
        if(toBeReplaced) {
            def replacement = [:]
            replacement[toBeReplaced] = calls[toBeReplaced] as Set
            def removed = hereWeNeedToReplace.remove(toBeReplaced)
            hereWeNeedToReplace.add(replacement)
            counter++
        } else {
            alreadyHandled << stepName
        }
        break
    }
}

if(! done) {
    throw new Exception('Unable to resolve transitive plugin calls.')
}

piperStepCallMappings = [:]

for(def entry : calls.entrySet()) {
    def performedCalls = flatten(entry, (Set)[])
    piperStepCallMappings.put(entry.key, performedCalls)
}

//
// special handling since since changeManagement util class
// is separated from the steps itself
//
// should be improved in the future in order not to have
// that bells and whistles here.

def cm = piperStepCallMappings.get('changeManagement')

for (cmStepName in [
    'checkChangeInDevelopment',
    'transportRequestCreate',
    'transportRequestUploadFile',
    'transportRequestRelease',
]) {
    piperStepCallMappings.get(cmStepName).addAll(cm)
}

// end of special handling
//

File performedCalls = new File(options.o)
if (performedCalls.exists()) performedCalls.delete()
performedCalls << groovy.json.JsonOutput.toJson(piperStepCallMappings)

def flatten(def entry, Set result) {

    for(def e : entry.value) {
        if(e in Map) { // the map here is expected to hold one entry always
            for(def steps : e.entrySet().value) {
                for(def step : steps) {
                    if (step in Map) {
                        flatten(step, result)
                    } else {
                        result << step
                    }
                }
            }
        } else {
            result << e.value.toString()
        }
    }
    result
}

static retrieveStepName(String s) {
    firstCharToLowerCase(removeTrailing(s, 'Test'))
}

static removeTrailing(String s, String trail) {
    return s.replaceAll(trail + '$', '')
}

static firstCharToLowerCase(CharSequence cs) {
    char[] c = cs.getChars()
    c[0] = Character.toLowerCase(c[0])
    new String(c)
}
