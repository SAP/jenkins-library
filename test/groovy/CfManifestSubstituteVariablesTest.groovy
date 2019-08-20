import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.junit.Assert.*
import static util.JenkinsWriteYamlRule.DATA
import static util.JenkinsWriteYamlRule.SERIALIZED_YAML

public class CfManifestSubstituteVariablesTest extends BasePiperTest {

    private JenkinsStepRule script = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsWriteYamlRule writeYamlRule = new JenkinsWriteYamlRule(this)
    private JenkinsErrorRule errorRule = new JenkinsErrorRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private ExpectedException expectedExceptionRule = ExpectedException.none()
    private JenkinsDeleteFileRule deleteFileRule = new JenkinsDeleteFileRule(this).skipDeletion(true)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(writeYamlRule)
        .around(errorRule)
        .around(environmentRule)
        .around(loggingRule)
        .around(script)
        .around(deleteFileRule)
        .around(expectedExceptionRule)

    @Before
    public void setup() {
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
    public void substituteVariables_SkipsExecution_If_ManifestNotPresent() throws Exception {
        String manifestFileName = "nonexistent/manifest.yml"
        String variablesFileName = "nonexistent/manifest-variables.yml"

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Could not find YAML file at ${manifestFileName}. Skipping variable substitution.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_SkipsExecution_If_VariablesFileNotPresent() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "nonexistent/manifest-variables.yml"

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Could not find variable substitution file at ${variablesFileName}. Skipping variable substitution.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_Throws_If_manifestInvalid() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/invalid_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/invalid_manifest.yml"

        //check that exception is thrown and that it has the correct message.
        expectedExceptionRule.expect(org.yaml.snakeyaml.scanner.ScannerException)
        expectedExceptionRule.expectMessage("found character '%' that cannot start any token. (Do not use % for indentation)")

        loggingRule.expect("[CFManifestSubstituteVariables] Could not load manifest file at ${manifestFileName}.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_Throws_If_manifestVariablesInvalid() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/invalid_manifest.yml"

        //check that exception is thrown and that it has the correct message.
        expectedExceptionRule.expect(org.yaml.snakeyaml.scanner.ScannerException)
        expectedExceptionRule.expectMessage("found character '%' that cannot start any token. (Do not use % for indentation)")

        loggingRule.expect("[CFManifestSubstituteVariables] Could not load manifest variables file at ${variablesFileName}")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_UsesDefaultFileName_If_NoManifestSpecified() throws Exception {
        // In this test, we check that the implementation will resort to the default manifest file name.
        // Since the file is not present, the implementation should stop, but the log should indicate that the
        // the default file name was used.

        String manifestFileName = "manifest.yml" // default name should be chosen.

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Could not find YAML file at ${manifestFileName}. Skipping variable substitution.")

        // execute step
        script.step.cfManifestSubstituteVariables script: nullScript, skipDeletion: true

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_UsesDefaultFileName_If_NoVariablesFileSpecified() throws Exception {
        // In this test, we check that the implementation will resort to the default manifest _variables_ file name.
        // Since the file is not present, the implementation should stop, but the log should indicate that the
        // the default file name was used.

        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "manifest-variables.yml" // default file name that should be chosen.

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Could not find variable substitution file at ${variablesFileName}. Skipping variable substitution.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, script: nullScript

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

        // check that the step was marked as a success (even if it did do nothing).
        assertJobStatusSuccess()
    }

    @Test
    public void substituteVariables_ReplacesVariablesProperly_InSingleYamlFiles() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Loaded manifest at ${manifestFileName}!")
                   .expect("[CFManifestSubstituteVariables] Loaded variables file at ${variablesFileName}!")
                   .expect("[CFManifestSubstituteVariables] Replaced variables in ${manifestFileName} with variables from ${variablesFileName}.")
                   .expect("[CFManifestSubstituteVariables] Wrote output file (with variables replaced) at ${manifestFileName}.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript


        String yamlStringAfterReplacement = writeYamlRule.files[manifestFileName].get(SERIALIZED_YAML) as String
        Map<String, Object> manifestDataAfterReplacement = writeYamlRule.files[manifestFileName].get(DATA)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        // check that there are no unresolved variables left.
        assertAllVariablesReplaced(yamlStringAfterReplacement)

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
    public void substituteVariables_ReplacesVariablesProperly_InMultiYamlFiles() throws Exception {
        String manifestFileName = "test/resources/variableSubstitution/multi_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Loaded manifest at ${manifestFileName}!")
                   .expect("[CFManifestSubstituteVariables] Loaded variables file at ${variablesFileName}!")
                   .expect("[CFManifestSubstituteVariables] Replaced variables in ${manifestFileName} with variables from ${variablesFileName}.")
                   .expect("[CFManifestSubstituteVariables] Wrote output file (with variables replaced) at ${manifestFileName}.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript


        String yamlStringAfterReplacement = writeYamlRule.files[manifestFileName].get(SERIALIZED_YAML) as String
        List<Object> manifestDataAfterReplacement = writeYamlRule.files[manifestFileName].get(DATA)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        // check that there are no unresolved variables left.
        assertAllVariablesReplaced(yamlStringAfterReplacement)

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
    public void substituteVariables_SkipsExecution_If_NoVariablesInManifest() throws Exception {
        // This test makes sure that, if no variables are found in a manifest that need
        // to be replaced, the execution is eventually skipped and the manifest remains
        // untouched.

        String manifestFileName = "test/resources/variableSubstitution/novars_manifest.yml"
        String variablesFileName = "test/resources/variableSubstitution/manifest-variables.yml"

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Loaded manifest at ${manifestFileName}!")
                   .expect("[CFManifestSubstituteVariables] Loaded variables file at ${variablesFileName}!")
                   .expect("[CFManifestSubstituteVariables] No variables were found or could be replaced in ${manifestFileName}. Skipping variable substitution.")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript

        //Check that nothing was written
        assertNull(writeYamlRule.files[manifestFileName])

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

        // check that a proper log is written.
        loggingRule.expect("[CFManifestSubstituteVariables] Loaded manifest at ${manifestFileName}!")
            .expect("[CFManifestSubstituteVariables] Loaded variables file at ${variablesFileName}!")

        // execute step
        script.step.cfManifestSubstituteVariables manifestFile: manifestFileName, variablesFile: variablesFileName, script: nullScript

        String yamlStringAfterReplacement = writeYamlRule.files[manifestFileName].get(SERIALIZED_YAML) as String
        Map<String, Object> manifestDataAfterReplacement = writeYamlRule.files[manifestFileName].get(DATA)

        //Check that something was written
        assertNotNull(manifestDataAfterReplacement)

        assertAllVariablesReplaced(yamlStringAfterReplacement)
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

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("booleanVariable").equals(true))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("booleanVariable") instanceof Boolean)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("floatVariable") == 0.25)
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("floatVariable") instanceof Double)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("json-variable") instanceof String)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("object-variable") instanceof Map)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("string-variable").startsWith("true-0.25-1-"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("string-variable") instanceof String)

        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("single-var-with-string-constants").equals("true-with-some-more-text"))
        assertTrue(manifestDataAfterReplacement.get("applications").get(0).get("env").get("single-var-with-string-constants") instanceof String)
    }
}
