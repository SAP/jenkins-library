package util

import static com.lesfurets.jenkins.unit.MethodSignature.method
import static util.StepHelper.getSteps

import org.codehaus.groovy.runtime.MetaClassHelper
import com.lesfurets.jenkins.unit.MethodSignature
import com.lesfurets.jenkins.unit.PipelineTestHelper
import groovy.json.JsonBuilder

class StepTracker {

    /*
     * Contains the piper steps as key (derived from the test name, so this is blurry since it might
     * contains also other cases than only piper step name) and the observed calls in a collection.
     */
    static Map piperStepCallMapping = [:]
    static Set piperSteps = StepHelper.getSteps()

    static Set calls

    static {
        initialize()
    }

    final static void initialize() {

        PipelineTestHelper.metaClass.getAllowedMethodEntry = {

            // We need to be careful here, in case we switch to another
            // version of the Les Furets framework we have to check if
            // this here still works.

            String name, Object[] args ->

            Class[] paramTypes = MetaClassHelper.castArgumentsToClassArray(args)
            MethodSignature signature = method(name, paramTypes)
            def intercepted = allowedMethodCallbacks.find { k, v -> k == signature }

            if(intercepted != null)
                StepTracker.add(name)

            return intercepted
        }
    }

    static void before(String stepName) {

        if(piperStepCallMapping[stepName] == null)
            piperStepCallMapping[stepName] = (Set)[]
        calls = piperStepCallMapping[stepName]
    }

    static void after() {
        calls = null
        write()
    }
    static void add (String call) {
        calls.add(call)
    }

    static private void write() {
        Map root = [
            piperSteps: piperSteps,
            calls: piperStepCallMapping.sort()
        ]
        new File('target/trackedCalls.json').write(new JsonBuilder(root).toPrettyString())
    }
}
