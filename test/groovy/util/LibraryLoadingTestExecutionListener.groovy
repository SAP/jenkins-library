package util

import com.lesfurets.jenkins.unit.MethodSignature
import com.lesfurets.jenkins.unit.PipelineTestHelper
import org.springframework.test.context.TestContext
import org.springframework.test.context.support.AbstractTestExecutionListener
import org.springframework.test.context.support.DependencyInjectionTestExecutionListener

import static com.lesfurets.jenkins.unit.MethodSignature.method

class LibraryLoadingTestExecutionListener extends AbstractTestExecutionListener {

    static PipelineTestHelper singletonInstance

    static List TRACKED_ON_CLASS = []
    static List TRACKED_ON_METHODS = []

    static HashMap RESTORE_ON_CLASS = [:]
    static HashMap RESTORE_ON_METHODS = [:]

    static boolean START_METHOD_TRACKING = false
    static boolean START_CLASS_TRACKING = false

    @Override
    int getOrder() {
        return 2500
    }

    static void setSingletonInstance(PipelineTestHelper helper) {
        if(null != helper) {
            helper.metaClass.invokeMethod = {
                String name, Object[] args ->
                    if ((LibraryLoadingTestExecutionListener.START_METHOD_TRACKING || LibraryLoadingTestExecutionListener.START_CLASS_TRACKING)
                        && name.equals("registerAllowedMethod")) {

                        List list
                        HashMap restore
                        if (LibraryLoadingTestExecutionListener.START_METHOD_TRACKING) {
                            list = LibraryLoadingTestExecutionListener.TRACKED_ON_METHODS
                            restore = LibraryLoadingTestExecutionListener.RESTORE_ON_METHODS
                        } else if (LibraryLoadingTestExecutionListener.START_CLASS_TRACKING) {
                            list = LibraryLoadingTestExecutionListener.TRACKED_ON_CLASS
                            restore = LibraryLoadingTestExecutionListener.RESTORE_ON_CLASS
                        }

                        Object methodName = args[0]
                        def key
                        if (args[1] instanceof List) {
                            List<Class> parameters = args[1]
                            key = method(methodName, parameters.toArray(new Class[parameters.size()]))
                        } else if (!(args[0] instanceof MethodSignature)) {
                            key = method(methodName, args[1])
                        }

                        if (null != key) {
                            list.add(key)
                            def existingValue = helper.removeAllowedMethodCallback(key)
                            if (!restore.containsKey(key)) {
                                restore.put(key, existingValue)
                            }
                        }
                    }
                    def m = delegate.metaClass.getMetaMethod(name, *args)
                    if( null != m)
                        return m.invoke(delegate, *args)
            }
        }
        singletonInstance = helper
    }

    static PipelineTestHelper getSingletonInstance() {
        if (singletonInstance == null) {
            setSingletonInstance(new LibraryLoadingTestExecutionListener.PipelineTestHelperHook().helper)
        }

        return singletonInstance
    }

    @Override
    void beforeTestClass(TestContext testContext) throws Exception {
        super.beforeTestClass(testContext)
        StepTracker.before(testContext.testClass.getSimpleName())
        def helper = LibraryLoadingTestExecutionListener.getSingletonInstance()
        registerDefaultAllowedMethods(helper)
        LibraryLoadingTestExecutionListener.START_CLASS_TRACKING = true
    }

    @Override
    void afterTestClass(TestContext testContext) throws Exception {
        super.afterTestClass(testContext)
        StepTracker.after()
        PipelineTestHelper helper = LibraryLoadingTestExecutionListener.getSingletonInstance()
        helper.clearAllowedMethodCallbacks(LibraryLoadingTestExecutionListener.TRACKED_ON_CLASS)
        LibraryLoadingTestExecutionListener.TRACKED_ON_CLASS.clear()

        helper.putAllAllowedMethodCallbacks(LibraryLoadingTestExecutionListener.RESTORE_ON_CLASS)
        LibraryLoadingTestExecutionListener.RESTORE_ON_CLASS.clear()

        LibraryLoadingTestExecutionListener.START_CLASS_TRACKING = false

        if (Boolean.TRUE.equals(testContext.getAttribute(DependencyInjectionTestExecutionListener.REINJECT_DEPENDENCIES_ATTRIBUTE))) {
            LibraryLoadingTestExecutionListener.singletonInstance = null
            PipelineTestHelper newHeiper = LibraryLoadingTestExecutionListener.getSingletonInstance()

            def applicationContext = testContext.getApplicationContext()
            def beanNames = applicationContext.getBeanDefinitionNames()
            beanNames.each { name ->
                LibraryLoadingTestExecutionListener.prepareObjectInterceptors(applicationContext.getBean(name))
            }
        }
    }

    @Override
    void beforeTestMethod(TestContext testContext) throws Exception {
        super.beforeTestMethod(testContext)
        def testInstance = testContext.getTestInstance()
        StepTracker.before(testInstance.getClass().getSimpleName())
        testInstance.binding.setVariable('currentBuild', [result: 'SUCCESS', currentResult: 'SUCCESS'])
        PipelineTestHelper helper = LibraryLoadingTestExecutionListener.getSingletonInstance()
        LibraryLoadingTestExecutionListener.START_METHOD_TRACKING = true
    }

    @Override
    void afterTestMethod(TestContext testContext) throws Exception {
        super.afterTestMethod(testContext)
        def testInstance = testContext.getTestInstance()
        StepTracker.after()
        PipelineTestHelper helper = LibraryLoadingTestExecutionListener.getSingletonInstance()

        helper.clearCallStack()
        helper.clearAllowedMethodCallbacks(LibraryLoadingTestExecutionListener.TRACKED_ON_METHODS)
        LibraryLoadingTestExecutionListener.TRACKED_ON_METHODS.clear()

        helper.putAllAllowedMethodCallbacks(LibraryLoadingTestExecutionListener.RESTORE_ON_METHODS)
        LibraryLoadingTestExecutionListener.RESTORE_ON_METHODS.clear()

        LibraryLoadingTestExecutionListener.START_METHOD_TRACKING = false

        testInstance.getNullScript().commonPipelineEnvironment.reset()
    }

    def registerDefaultAllowedMethods(helper) {
        helper.registerAllowedMethod("stage", [String.class, Closure.class], null)
        helper.registerAllowedMethod("stage", [String.class, Closure.class], null)
        helper.registerAllowedMethod("node", [String.class, Closure.class], null)
        helper.registerAllowedMethod("node", [Closure.class], null)
        helper.registerAllowedMethod( method('sh', Map.class), {m ->
            return ""
        } )
        helper.registerAllowedMethod( method('sh', String.class), {s ->
            return ""
        } )
        helper.registerAllowedMethod("checkout", [Map.class], null)
        helper.registerAllowedMethod("echo", [String.class], null)
        helper.registerAllowedMethod("timeout", [Map.class, Closure.class], null)
        helper.registerAllowedMethod("step", [Map.class], null)
        helper.registerAllowedMethod("input", [String.class], null)
        helper.registerAllowedMethod("gitlabCommitStatus", [String.class, Closure.class], { String name, Closure c ->
            c.delegate = delegate
            helper.callClosure(c)
        })
        helper.registerAllowedMethod("gitlabBuilds", [Map.class, Closure.class], null)
        helper.registerAllowedMethod("logRotator", [Map.class], null)
        helper.registerAllowedMethod("buildDiscarder", [Object.class], null)
        helper.registerAllowedMethod("pipelineTriggers", [List.class], null)
        helper.registerAllowedMethod("properties", [List.class], null)
        helper.registerAllowedMethod("dir", [String.class, Closure.class], null)
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], null)
        helper.registerAllowedMethod("junit", [String.class], null)
        helper.registerAllowedMethod("readFile", [String.class], null)
        helper.registerAllowedMethod("disableConcurrentBuilds", [], null)
        helper.registerAllowedMethod("gatlingArchive", [], null)

        helper.registerAllowedMethod("unstash", [String.class], null)
        helper.registerAllowedMethod("unstash", [Object.class, String.class], null)
        helper.registerAllowedMethod("stash", [Map.class], null)
        helper.registerAllowedMethod("echo", [String.class], null)
    }

    static def prepareObjectInterceptors(Object object) {
        object.metaClass.invokeMethod = LibraryLoadingTestExecutionListener.singletonInstance.getMethodInterceptor()
        object.metaClass.static.invokeMethod = LibraryLoadingTestExecutionListener.singletonInstance.getMethodInterceptor()
        object.metaClass.methodMissing = LibraryLoadingTestExecutionListener.singletonInstance.getMethodMissingInterceptor()
    }

    static class PipelineTestHelperHook {
        def helper = new PipelineTestHelper() {

            def clearAllowedMethodCallbacks(Collection c = []) {
                List itemsToRemove = []
                c.each {
                    key ->
                        allowedMethodCallbacks.entrySet().each {
                            entry ->
                                if (entry?.getKey().equals(key))
                                    itemsToRemove.add(entry.getKey())
                        }
                }
                allowedMethodCallbacks.keySet().removeAll(itemsToRemove)
            }

            def removeAllowedMethodCallback(Object key) {
                def itemToRemove
                allowedMethodCallbacks.entrySet().each {
                    entry ->
                        if (entry?.getKey().equals(key)) {
                            itemToRemove = entry.getKey()
                        }
                }
                if (null != itemToRemove) {
                    def itemValue = allowedMethodCallbacks.remove(itemToRemove)
                    return itemValue
                }
                return null
            }

            def putAllAllowedMethodCallbacks(HashMap m) {
                m.entrySet().each {
                    entry ->
                        if(null != entry.getValue())
                            allowedMethodCallbacks.put(entry.getKey(), entry.getValue())
                        else
                            allowedMethodCallbacks.remove(entry.getKey())
                }
            }

            protected void setGlobalVars(Binding binding) {
                libLoader.libRecords.values().stream()
                    .flatMap { it.definedGlobalVars.entrySet().stream() }
                    .forEach { e ->
                    if (e.value instanceof Script) {
                        Script script = Script.cast(e.value)
                        // invoke setBinding from method to avoid interception
                        SCRIPT_SET_BINDING.invoke(script, binding)
                        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(script)
                        script.metaClass.getMethods().findAll { it.name == 'call' }.forEach { m ->
                            LibraryLoadingTestExecutionListener.singletonInstance.registerAllowedMethod(method(e.value.class.name, m.getNativeParameterTypes()),
                                { args ->
                                    m.doMethodInvoke(e.value, args)
                                })
                        }
                    }
                    binding.setVariable(e.key, e.value)
                }
            }
        }
    }
}
