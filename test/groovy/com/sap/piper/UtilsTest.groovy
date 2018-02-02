package com.sap.piper

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.Rules
import util.SharedLibraryCreator

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasSize
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat

class UtilsTest extends BasePipelineTest {

    @Rule
    public ExpectedException exception = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules.getCommonRules(this, SharedLibraryCreator.lazyLoadedLibrary)

    Utils utils

    @Before
    void init() throws Exception {
        utils = new Utils()
        prepareObjectInterceptors(utils)
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }

    @Test
    void testGetMandatoryParameterValid() {

        def sourceMap = [test1: 'value1', test2: 'value2']

        def defaultFallbackMap = [myDefault1: 'default1']

        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', null))

        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', ''))

        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', 'customValue'))

    }

    @Test
    void testGetMandatoryParameterDefaultFallback() {

        def myMap = [test1: 'value1', test2: 'value2']

        assertEquals('', utils.getMandatoryParameter(myMap, 'test3', ''))
        assertEquals('customValue', utils.getMandatoryParameter(myMap, 'test3', 'customValue'))
    }


    @Test
    void testGetMandatoryParameterFail() {

        def myMap = [test1: 'value1', test2: 'value2']

        exception.expect(Exception.class)

        exception.expectMessage("ERROR - NO VALUE AVAILABLE FOR")

        utils.getMandatoryParameter(myMap, 'test3', null)
    }
}
