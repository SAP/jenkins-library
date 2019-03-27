import groovy.json .JsonSlurper

def steps = new JsonSlurper().parseText(new File('target/trackedCalls.json').text)

def piperSteps = steps.piperSteps
def calls = steps.calls

def _calls = [:]

// Adjust naming
calls.each { c ->
    _calls.put(retrieveStepName(c.key), c.value as Set)
}

calls = _calls
_calls = null

// Remove selfs
calls.each { c ->
    c.value.remove(c.key)
} 

int counter=0

def alreadyHandled = []

while(counter < 1600) {

    def hereWeNeedToReplace
    def toBeReplaced

    if(alreadyHandled.size() == calls.size()) break

    for (def call in calls.entrySet()) {

        stepName = call.key
        calledSteps = call.value

        if(alreadyHandled.contains(stepName)) {
            continue
        }

        for (def calledStep in calledSteps) {

            if (calledStep in Map) {
            } else {
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


piperStepCallMappings = [:]

for(def entry : calls.entrySet()) {
    def performedCalls = flatten(entry, (Set)[])
    piperStepCallMappings.put(entry.key, performedCalls)
}

File performedCalls = new File('target/performedCalls.json')
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
