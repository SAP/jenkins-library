package com.sap.piper.tools.neo

import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsFileExistsRule
import util.Rules

class NeoCommandHelperTest extends BasePiperTest {


    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, ['file.mta', 'file.war', 'file.properties'])
    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(fileExistsRule)

    NeoCommandHelper getTestFixture(DeployMode deployMode, Set extensions = []) {

        Map deploymentConfiguration = [
            host          : 'host_value',
            account       : 'account_value',
            application   : 'application_value',
            environment   : [ENV1: 'value1', ENV2: 'value2'],
            vmArguments   : '-Dargument1=value1 -Dargument2=value2',
            runtime       : 'neо-javaee6-wp',
            runtimeVersion: '2',
            size          : 'lite',
            propertiesFile: 'file.properties',
            azDistribution: '2'
        ]

        String source = (deployMode == DeployMode.MTA) ? 'file.mta' : 'file.war'
        String username = 'username'
        String password = 'password'

        nullScript.STEP_NAME="neoDeploy"

        return new NeoCommandHelper(
            nullScript,
            deployMode,
            deploymentConfiguration,
            extensions,
            username,
            password,
            source
        )
    }

    @Test
    void testStatusCommand() {
        String actual = getTestFixture(DeployMode.WAR_PARAMS).statusCommand()
        String expected = "neo.sh status --host 'host_value' --account 'account_value' " +
            "--application 'application_value' --user 'username' --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testStatusCommandForProperties() {
        String actual = getTestFixture(DeployMode.WAR_PROPERTIES_FILE).statusCommand()
        String expected = "neo.sh status file.properties --user 'username' --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testRollingUpdateCommand() {
        String actual = getTestFixture(DeployMode.WAR_PARAMS).rollingUpdateCommand()
        String basicCommand = "neo.sh rolling-update --host 'host_value' --account 'account_value' " +
            "--application 'application_value' --user 'username' --password 'password' --source 'file.war'"

        Assert.assertTrue(actual.contains(basicCommand))
        Assert.assertTrue(actual.contains(' --ev \'ENV1\'=\'value1\' --ev \'ENV2\'=\'value2\''))
        Assert.assertTrue(actual.contains(' --vm-arguments \'-Dargument1=value1 -Dargument2=value2\''))
        Assert.assertTrue(actual.contains('--runtime \'neо-javaee6-wp\''))
        Assert.assertTrue(actual.contains(' --runtime-version \'2\''))
        Assert.assertTrue(actual.contains(' --size \'lite\''))
        Assert.assertTrue(actual.contains(' --az-distribution \'2\''))
    }

    @Test
    void testRollingUpdateCommandForProperties() {
        String actual = getTestFixture(DeployMode.WAR_PROPERTIES_FILE).rollingUpdateCommand()
        String expected = "neo.sh rolling-update file.properties --user 'username' --password 'password' --source 'file.war' "
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testDeployCommand() {
        String actual = getTestFixture(DeployMode.WAR_PARAMS).deployCommand()
        String basicCommand = "neo.sh deploy --host 'host_value' --account 'account_value' " +
            "--application 'application_value' --user 'username' --password 'password' --source 'file.war'"

        Assert.assertTrue(actual.contains(basicCommand))
        Assert.assertTrue(actual.contains(' --ev \'ENV1\'=\'value1\' --ev \'ENV2\'=\'value2\''))
        Assert.assertTrue(actual.contains(' --vm-arguments \'-Dargument1=value1 -Dargument2=value2\''))
        Assert.assertTrue(actual.contains(' --runtime \'neо-javaee6-wp\''))
        Assert.assertTrue(actual.contains(' --runtime-version \'2\''))
        Assert.assertTrue(actual.contains(' --size \'lite\''))
        Assert.assertTrue(actual.contains(' --az-distribution \'2\''))
    }

    @Test
    void testDeployCommandForProperties() {
        String actual = getTestFixture(DeployMode.WAR_PROPERTIES_FILE).deployCommand()
        String expected = "neo.sh deploy file.properties --user 'username' --password 'password' --source 'file.war' "
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testRestartCommand() {
        String actual = getTestFixture(DeployMode.WAR_PARAMS).restartCommand()
        String expected = "neo.sh restart --synchronous --host 'host_value' --account 'account_value' " +
            "--application 'application_value' --user 'username' --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void testRestartCommandForProperties() {
        String actual = getTestFixture(DeployMode.WAR_PROPERTIES_FILE).restartCommand()
        String expected = "neo.sh restart --synchronous file.properties --user 'username' --password 'password'"
        Assert.assertEquals(expected, actual)
    }

    @Test
    void deployMta() {
        String actual = getTestFixture(DeployMode.MTA, (Set)['myExtension1.yml', 'myExtension2.yml']).deployMta()
        String expected = "neo.sh deploy-mta --synchronous --host 'host_value' --account 'account_value' " +
            "--user 'username' --password 'password' --extensions 'myExtension1.yml','myExtension2.yml' --source 'file.mta'"
        Assert.assertEquals(expected, actual)
    }
}
