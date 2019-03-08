package util

import groovy.json.JsonBuilder

class StepTracker {

        final static filePath = 'target/trackedCalls.json'

    final static trackedCalls = [:]

    def add(String testClass, Collection steps) {

        def stepName = stepName(testClass)

        if(trackedCalls[stepName] == null)
            trackedCalls[stepName] = (Set)[]

        steps.each { c -> trackedCalls[stepName] << c.name }

        persist()
    }

    private persist() {
        new File(filePath).write(new JsonBuilder(trackedCalls).toPrettyString())
    }

    // we expect a naming convention between test class and step under test:
    // "<step>Test", where the first char of the test is transformed to upper case.
    private static stepName(CharSequence testClass) {
        firstCharToLowerCase(stripTrailingTest(testClass))
    }

    private static stripTrailingTest(CharSequence testClass) {
        testClass.replaceAll('Test$', '')
    }

    private static firstCharToLowerCase(CharSequence cs) {
        char[] c = cs.getChars()
        c[0] = Character.toLowerCase(c[0])
        return new String(c)
    }
}
