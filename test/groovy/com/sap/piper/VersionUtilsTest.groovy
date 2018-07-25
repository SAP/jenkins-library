package com.sap.piper

import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsShellCallRule
import util.Rules

import static org.junit.Assert.assertEquals
import static org.hamcrest.Matchers.equalTo
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertFalse
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.notNullValue
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertNull
import static org.junit.Assert.assertThat

class VersionUtilsTest extends BasePiperTest {

    ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this).around(thrown)

    @Before
    void init() throws Exception {
    }

    @Test
    void test_if_getVersionDesc_returns_desc() {
        helper.registerAllowedMethod('sh', [Map], { Map m -> return 'SAP Cloud Platform Console Client\n\n\nSDK version    : 2.129.5.1\nRuntime        : neo-javaee6-wp\n'}) 
       
        assertEquals('SAP Cloud Platform Console Client\n\n\nSDK version    : 2.129.5.1\nRuntime        : neo-javaee6-wp\n',VersionUtils.getVersionDesc(nullScript, "test", "test.sh", "version"))
    }

    @Test
    void test_if_getVersion_returns_version() {
        helper.registerAllowedMethod('sh', [Map], { Map m -> return 'SAP Cloud Platform Console Client\n\n\nSDK version    : 2.129.5.1\nRuntime        : neo-javaee6-wp\n'}) 
       
        assertEquals(new Version('2.129.5.1'),VersionUtils.getVersion(nullScript, "test", "test.sh", "version"))
    }

    @Test
    void test_if_verifyVersion_succeeds_compatible() {
        helper.registerAllowedMethod('sh', [Map], { Map m -> return 'version : 1.0.0\runtime: key' })
        VersionUtils.verifyVersion(nullScript, "test", "test.sh", '1.0.0', "version")
    }
    
    @Test
    void test_if_verifyVersion_fails_incompatible() {
        helper.registerAllowedMethod('sh', [Map], { Map m -> return 'version : 1.0.0\runtime: key' })
        
        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of test is 1.0.0. Please install version 1.0.1 or a compatible version.")

        VersionUtils.verifyVersion(nullScript, "test", "test.sh", '1.0.1', "version")
    }

    @Test
    void test_if_verifyVersion_map_succeeds_compatible() {
        Map versionMap = ['key1': '1.0.0', 'key2': '2.0.0', 'key3': '3.0.0']
        helper.registerAllowedMethod('sh', [Map], { Map m -> return 'version : 1.0.0\runtime: key1' })
        VersionUtils.verifyVersion(nullScript, "test", "test.sh", versionMap, "version")
    }
    
    @Test
    void test_if_verifyVersion_map_fails_incompatible() {
        Map versionMap = ['key1': '1.0.1', 'key2': '2.0.1', 'key3': '3.0.1']
        helper.registerAllowedMethod('sh', [Map], { Map m -> return 'version : 1.0.0\runtime: key1' })
        
        thrown.expect(AbortException)
        thrown.expectMessage("The installed version of test is 1.0.0. Please install version 1.0.1 or a compatible version.")
        
        VersionUtils.verifyVersion(nullScript, "test", "test.sh", versionMap, "version")
    }
    

}
