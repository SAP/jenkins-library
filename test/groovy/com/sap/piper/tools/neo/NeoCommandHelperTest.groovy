package com.sap.piper.tools.neo


import org.junit.Assert
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsErrorRule
import util.JenkinsFileExistsRule
import util.Rules

class NeoCommandHelperTest extends BasePiperTest {


    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, ['file.mta', 'file.war', 'file.properties'])
    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(fileExistsRule)

    NeoCommandHelper getTestFixture(String deployMode) {

        Map deploymentConfiguration = [
            host          : 'host_value',
            account       : 'account_value',
            application   : 'application_value',
            environment   : [ENV1: 'value1', ENV2: 'value2'],
            vmArguments   : '-Dargument1=value1 -Dargument2=value2',
            runtime       : 'neо-javaee6-wp',
            runtimeVersion: '2',
            size          : 'lite',
            propertiesFile: 'file.properties'
        ]

        String source = (deployMode == 'mta') ?'file.mta' :'file.war'
        String username = 'username'
        String password = 'password'
        String neoExecutable = '/path/tools/neo.sh';

        return new NeoCommandHelper(
            nullScript,
            deployMode,
            deploymentConfiguration,
            neoExecutable,
            username,
            password,
            source
        )
    }

    @Test
    void testStatusCommand() {
        String actual = getTestFixture('warParams').statusCommand()
        String expected = "/path/tools/neo.sh status --host host_value --account account_value " +
            "--application application_value --username username --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testStatusCommandForProperties() {
        String actual = getTestFixture('warPropertiesFile').statusCommand()
        String expected = "/path/tools/neo.sh status file.properties --username username --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testRollingUpdateCommand() {
        String actual = getTestFixture('warParams').rollingUpdateCommand()
        String basicCommand = "/path/tools/neo.sh rolling-update --host host_value --account account_value " +
            "--application application_value --username username --password 'password' --source file.war"

        Assert.assertTrue(actual.contains(basicCommand))
        Assert.assertTrue(actual.contains(' --ev \'ENV1\'=\'value1\' --ev \'ENV2\'=\'value2\''))
        Assert.assertTrue(actual.contains(' --vm-arguments "-Dargument1=value1 -Dargument2=value2"'))
        Assert.assertTrue(actual.contains('--runtime neо-javaee6-wp'))
        Assert.assertTrue(actual.contains(' --runtime-version 2'))
        Assert.assertTrue(actual.contains(' --size lite'))
    }

    @Test
    void testRollingUpdateCommandForProperties() {
        String actual = getTestFixture('warPropertiesFile').rollingUpdateCommand()
        String expected = "/path/tools/neo.sh rolling-update file.properties --username username --password 'password' --source file.war "
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testDeployCommand() {
        String actual = getTestFixture('warParams').deployCommand()
        String basicCommand = "/path/tools/neo.sh deploy --host host_value --account account_value " +
            "--application application_value --username username --password 'password' --source file.war"

        Assert.assertTrue(actual.contains(basicCommand))
        Assert.assertTrue(actual.contains(' --ev \'ENV1\'=\'value1\' --ev \'ENV2\'=\'value2\''))
        Assert.assertTrue(actual.contains(' --vm-arguments "-Dargument1=value1 -Dargument2=value2"'))
        Assert.assertTrue(actual.contains(' --runtime neо-javaee6-wp'))
        Assert.assertTrue(actual.contains(' --runtime-version 2'))
        Assert.assertTrue(actual.contains(' --size lite'))
    }

    @Test
    void testDeployCommandForProperties() {
        String actual = getTestFixture('warPropertiesFile').deployCommand()
        String expected = "/path/tools/neo.sh deploy file.properties --username username --password 'password' --source file.war "
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testRestartCommand() {
        String actual = getTestFixture('warParams').restartCommand()
        String expected = "/path/tools/neo.sh restart --synchronous --host host_value --account account_value " +
            "--application application_value --username username --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testRestartCommandForProperties() {
        String actual = getTestFixture('warPropertiesFile').restartCommand()
        String expected = "/path/tools/neo.sh restart --synchronous file.properties --username username --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void deployMta() {
        String actual = getTestFixture('mta').deployMta()
        String expected = "/path/tools/neo.sh deploy-mta --synchronous --host host_value --account account_value " +
            "--username username --password 'password' --source file.mta"
        Assert.assertEquals(expected, actual)
    }
}
