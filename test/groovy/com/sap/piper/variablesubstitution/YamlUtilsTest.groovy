package com.sap.piper.variablesubstitution

import org.junit.Before

import static org.junit.Assert.*
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException;
import org.junit.rules.RuleChain;
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsErrorRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsWriteYamlRule
import util.Rules

class YamlUtilsTest extends BasePiperTest {

    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsWriteYamlRule writeYamlRule = new JenkinsWriteYamlRule(this)
    private JenkinsErrorRule errorRule = new JenkinsErrorRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private ExpectedException expectedExceptionRule = ExpectedException.none()

    private YamlUtils yamlUtils

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(writeYamlRule)
        .around(errorRule)
        .around(environmentRule)
        .around(loggingRule)
        .around(expectedExceptionRule)

    @Before
    public void setup() {
        yamlUtils = new YamlUtils(nullScript)

        readYamlRule.registerYaml("manifest.yml", new FileInputStream(new File("test/resources/variableSubstitution/manifest.yml")))
                    .registerYaml("manifest-variables.yml", new FileInputStream(new File("test/resources/variableSubstitution/manifest-variables.yml")))
                    .registerYaml("test/resources/variableSubstitution/manifest.yml", new FileInputStream(new File("test/resources/variableSubstitution/manifest.yml")))
                    .registerYaml("test/resources/variableSubstitution/manifest-variables.yml", new FileInputStream(new File("test/resources/variableSubstitution/manifest-variables.yml")))
                    .registerYaml("test/resources/variableSubstitution/invalid_manifest.yml", new FileInputStream(new File("test/resources/variableSubstitution/invalid_manifest.yml")))
                    .registerYaml("test/resources/variableSubstitution/novars_manifest.yml", new FileInputStream(new File("test/resources/variableSubstitution/novars_manifest.yml")))
                    .registerYaml("test/resources/variableSubstitution/multi_manifest.yml", new FileInputStream(new File("test/resources/variableSubstitution/multi_manifest.yml")))
                    .registerYaml("test/resources/variableSubstitution/datatypes_manifest.yml", new FileInputStream(new File("test/resources/variableSubstitution/datatypes_manifest.yml")))
                    .registerYaml("test/resources/variableSubstitution/datatypes_manifest-variables.yml", new FileInputStream(new File("test/resources/variableSubstitution/datatypes_manifest-variables.yml")))
    }

    @Test
    public void substituteVariables_Fails_If_InputYamlIsNullOrEmpty() throws Exception {

        expectedExceptionRule.expect(IllegalArgumentException)
        expectedExceptionRule.expectMessage("[YamlUtils] Input Yaml data must not be null or empty.")

        yamlUtils.substituteVariables(null, null)
    }

    @Test
    public void substituteVariables_Fails_If_VariablesYamlIsNullOrEmpty() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"

        expectedExceptionRule.expect(IllegalArgumentException)
        expectedExceptionRule.expectMessage("[YamlUtils] Variables Yaml data must not be null or empty.")

        Object input = nullScript.readYaml file: manifestFileName

        // execute step
        yamlUtils.substituteVariables(input, null)
    }

    @Test
    public void substituteVariables_Throws_If_InputYamlIsInvalid() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/invalid_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/invalid_manifest.yml"

        //check that exception is thrown and that it has the correct message.
        expectedExceptionRule.expect(org.yaml.snakeyaml.scanner.ScannerException)
        expectedExceptionRule.expectMessage("found character '%' that cannot start any token. (Do not use % for indentation)")

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        yamlUtils.substituteVariables(input, variables)
    }

    @Test
    public void substituteVariables_Throws_If_VariablesYamlInvalid() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/invalid_manifest.yml"

        //check that exception is thrown and that it has the correct message.
        expectedExceptionRule.expect(org.yaml.snakeyaml.scanner.ScannerException)
        expectedExceptionRule.expectMessage("found character '%' that cannot start any token. (Do not use % for indentation)")

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        yamlUtils.substituteVariables(input, variables)
    }

    @Test
    public void substituteVariables_ReplacesVariablesProperly_InSingleYamlFiles() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        Map<String, Object> manifestDataAfterReplacement = yamlUtils.substituteVariables(input, variables)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        // check that resolved variables have expected values
        assertCorrectVariableResolution(manifestDataAfterReplacement)

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    private void assertAllVariablesReplaced(String yamlStringAfterReplacement) {
        assertFalse(yamlStringAfterReplacement.contains("(("))
        assertFalse(yamlStringAfterReplacement.contains("))"))
    }

    private void assertCorrectVariableResolution(Map<String, Object> manifestDataAfterReplacement) {
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("name").equals("uniquePrefix-catalog-service-odatav2-0.0.1"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("routes").get(0).get("route").equals("uniquePrefix-catalog-service-odatav2-001.cfapps.eu10.hana.ondemand.com"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("services").get(0).equals("uniquePrefix-catalog-service-odatav2-xsuaa"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("services").get(1).equals("uniquePrefix-catalog-service-odatav2-hana"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("xsuaa-instance-name").equals("uniquePrefix-catalog-service-odatav2-xsuaa"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("db_service_instance_name").equals("uniquePrefix-catalog-service-odatav2-hana"))
    }

    @Test
    public void substituteVariables_ReplacesVariablesProperly_InMultiYamlData() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/multi_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        List<Object> manifestDataAfterReplacement = yamlUtils.substituteVariables(input, variables)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        //check that result still is a multi-YAML file.
        assertEquals("Dumped YAML after replacement should still be a multi-YAML file.",2, manifestDataAfterReplacement.size())

        // check that resolved variables have expected values
        manifestDataAfterReplacement.each { yaml ->
            assertCorrectVariableResolution(yaml as Map<String, Object>)
        }

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_ReturnsOriginalIfNoVariablesPresent() throws Exception {
        // This test makes sure that, if no variables are found in a manifest that need
        // to be replaced, the execution is eventually skipped and the manifest remains
        // untouched.

        String manifestFileName = "test/resources/variableSubstitution/novars_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        ExecutionContext context = new ExecutionContext()
        Object result = yamlUtils.substituteVariables(input, variables, context)

        //Check that nothing was written
        assertNotNull(result)
        assertFalse(context.variablesReplaced)

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_SupportsAllDataTypes() throws Exception {
        // This test makes sure that, all datatypes supported by YAML are also
        // properly substituted by the substituteVariables step.
        // In particular this includes variables of type:
        // Integer, Boolean, String, Float and inline JSON documents (which are parsed as multi-line strings)
        // and complex types (like other YAML objects).
        // The test also checks the differing behaviour when substituting nodes that only consist of a
        // variable reference and nodes that contains several variable references or additional string constants.

        String manifestFileName = "test/resources/variableSubstitution/datatypes_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/datatypes_manifest-variables.yml"

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        ExecutionContext context = new ExecutionContext()
        Map<String, Object> manifestDataAfterReplacement = yamlUtils.substituteVariables(input, variables, context)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        assertCorrectVariableResolution(manifestDataAfterReplacement)

        assertDataTypeAndSubstitutionCorrectness(manifestDataAfterReplacement)

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    private void assertDataTypeAndSubstitutionCorrectness(Map<String, Object> manifestDataAfterReplacement) {
        // See datatypes_manifest.yml and datatypes_manifest-variables.yml.
        // Note: For debugging consider turning on YAML writing to a file in JenkinsWriteYamlRule to see the
        // actual outcome of replacing variables (for visual inspection).

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("instances").equals(1))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("instances") instanceof Integer)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("services").get(0) instanceof String)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("booleanVariableTrue").equals(true))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("booleanVariableTrue") instanceof Boolean)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("booleanVariableFalse").equals(false))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("booleanVariableFalse") instanceof Boolean)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("floatVariable") == 0.25)
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("floatVariable") instanceof Double)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("json-variable") instanceof String)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("object-variable") instanceof Map)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("string-variable").startsWith("true-0.25-1-"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("string-variable") instanceof String)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("single-var-with-string-constants").equals("true-with-some-more-text"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("single-var-with-string-constants") instanceof String)
    }

    @Test
    public void substituteVariables_DoesNotFail_If_ExecutionContextIsNull() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        Object input = nullScript.readYaml file: manifestFileName
        Object variables = nullScript.readYaml file: variablesFileName

        // execute step
        Map<String, Object> manifestDataAfterReplacement = yamlUtils.substituteVariables(input, variables, null)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        // check that resolved variables have expected values
        assertCorrectVariableResolution(manifestDataAfterReplacement)

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }
}
