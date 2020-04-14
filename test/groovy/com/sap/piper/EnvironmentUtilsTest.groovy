package com.sap.piper

import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

class EnvironmentUtilsTest {

    @Test
    void testCxServerDirectoryExists() {
        Assert.assertFalse(EnvironmentUtils.cxServerDirectoryExists())
    }

    @Test
    void testGetDockerFile() {
        String cxServerCfgContents = '''
#---------------------------------------------#
#-- Build server configuration ---------------#
#---------------------------------------------#

#>> Address of the used docker registry. Override if you do not want to use Docker's default registry.
docker_registry=some_registry

#>> Name of the used docker image
docker_image="some_path/some_file"

#>> Enable TLS encryption
tls_enabled=true
'''
        Assert.assertEquals( 'some_path/some_file',
            EnvironmentUtils.getDockerFile(cxServerCfgContents))
    }

    @Test
    void testGetDockerFileSpaces() {
        String cxServerCfgContents = '''
#>> Name of the used docker image
 docker_image = "some_path/some_file:latest"
'''
        Assert.assertEquals( 'some_path/some_file:latest',
            EnvironmentUtils.getDockerFile(cxServerCfgContents))
    }

    @Test
    void testGetDockerFileTrailingSpaces() {
        String cxServerCfgContents = '''
#>> Name of the used docker image
docker_image="some_path/some_file:latest"
'''
        Assert.assertEquals( 'some_path/some_file:latest',
            EnvironmentUtils.getDockerFile(cxServerCfgContents))
    }

    @Test
    void testGetDockerFileSpacesInPath() {
        String cxServerCfgContents = '''
#>> Name of the used docker image
docker_image = "some path/some file:latest"
'''
        Assert.assertEquals( 'some path/some file:latest',
            EnvironmentUtils.getDockerFile(cxServerCfgContents))
    }
}
