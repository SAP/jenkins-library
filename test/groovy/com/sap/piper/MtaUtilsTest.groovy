package com.sap.piper

import org.junit.Before
import org.junit.ClassRule
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.junit.rules.TemporaryFolder
import org.yaml.snakeyaml.Yaml

import groovy.json.JsonSlurper
import hudson.AbortException
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.Rules


class MtaUtilsTest extends BasePiperTest {
    private static srcPackageJson = 'test/resources/MtaUtils/package.json'
    private static mtaTemplate = 'resources/template_mta.yaml'
    private static data
    private static String generatedFile
    private static String targetMtaDescriptor
    private File badJson
    private mtaUtils

    private ExpectedException thrown= ExpectedException.none()

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)

    @Before
    void init() {
        targetMtaDescriptor = "${tmp.getRoot()}/generated_mta.yml"
        mtaUtils = new MtaUtils(nullScript)

        this.helper.registerAllowedMethod('readJSON', [Map], { Map parameters ->
            return new JsonSlurper().parse(new File(parameters.file))
        })

        this.helper.registerAllowedMethod('libraryResource', [Map], {  Map parameters ->
            new Yaml().load(new File(mtaTemplate).text)
        })

        this.helper.registerAllowedMethod('readYaml', [], {
            return new Yaml().load(new FileReader(mtaTemplate))
        })

        this.helper.registerAllowedMethod('writeYaml', [Map], { Map parameters ->
            generatedFile = parameters.file
            data = parameters.data
        })

        this.helper.registerAllowedMethod('fileExists', [String.class], { true })
    }

    @Test
    void testStraightForward(){
        mtaUtils.generateMtaDescriptorFromPackageJson(srcPackageJson, targetMtaDescriptor, 'testAppName')
        assert data.ID == 'com.mycompany.northwind'
        assert data.version == '1.0.3'
        assert data.modules.name[0] == 'testAppName'
        assert data.modules.parameters.version[0] == '1.0.3-${timestamp}'
        assert data.modules.parameters.name[0] == 'testAppName'
    }

    @Test
    void testSrcPackageJsonEmpty() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'srcPackageJson' can not be null or empty.")
        mtaUtils.generateMtaDescriptorFromPackageJson('', targetMtaDescriptor, 'testApplicationName')
    }

    @Test
    void testSrcPackageJsonNull() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'srcPackageJson' can not be null or empty.")
        mtaUtils.generateMtaDescriptorFromPackageJson(null, targetMtaDescriptor, 'testApplicationName')
    }

    @Test
    void testTargetMtaDescriptorEmpty() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'targetMtaDescriptor' can not be null or empty.")
        mtaUtils.generateMtaDescriptorFromPackageJson(srcPackageJson, '', 'testApplicationName')
    }

    @Test
    void testTargetMtaDescriptorNull() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'targetMtaDescriptor' can not be null or empty.")
        mtaUtils.generateMtaDescriptorFromPackageJson(srcPackageJson, null, 'testApplicationName')
    }

    @Test
    void testApplicationNameEmpty() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'applicationName' can not be null or empty.")
        mtaUtils.generateMtaDescriptorFromPackageJson(srcPackageJson, targetMtaDescriptor, '')
    }

    @Test
    void testApplicationNameNull() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'applicationName' can not be null or empty.")
        mtaUtils.generateMtaDescriptorFromPackageJson(srcPackageJson, targetMtaDescriptor, null)
    }

    @Test
    void testMissingNameInJson() {
        badJson = tmp.newFile('missingName.json')
        badJson.text = missingNameInJson()
        badJson.dump()

        thrown.expect(AbortException)
        thrown.expectMessage("'name' not set in the given package.json.")

        mtaUtils.generateMtaDescriptorFromPackageJson(badJson.absolutePath, targetMtaDescriptor, 'testApplicationName')
    }

    @Test
    void testEmptyNameInJson() {
        badJson = tmp.newFile('emptyName.json')
        badJson.text = emptyNameInJson()
        badJson.dump()

        thrown.expect(AbortException)
        thrown.expectMessage("'name' not set in the given package.json.")

        mtaUtils.generateMtaDescriptorFromPackageJson(badJson.absolutePath, targetMtaDescriptor, 'testApplicationName')
    }

    @Test
    void testMissingVersionInJson() {
        badJson = tmp.newFile('missingVersion.json')
        badJson.text = missingVersionInJson()
        badJson.dump()

        thrown.expect(AbortException)
        thrown.expectMessage("'version' not set in the given package.json.")

        mtaUtils.generateMtaDescriptorFromPackageJson(badJson.absolutePath, targetMtaDescriptor, 'testApplicationName')
    }

    @Test
    void testEmptyVersionInJson() {
        badJson = tmp.newFile('emptyVersion.json')
        badJson.text = emptyVersionInJson()
        badJson.dump()

        thrown.expect(AbortException)
        thrown.expectMessage("'version' not set in the given package.json.")

        mtaUtils.generateMtaDescriptorFromPackageJson(badJson.absolutePath, targetMtaDescriptor, 'testApplicationName')
    }

    @Test
    void testFileGenerated() {
        mtaUtils.generateMtaDescriptorFromPackageJson(srcPackageJson, targetMtaDescriptor, 'testApplicationName')
        assert generatedFile.equals(targetMtaDescriptor)
    }


    private missingNameInJson() {
        return  '''
                {
				  "version": "1.0.3",
				  "description": "Webshop application for test purposes",
				  "private": true,
				  "devDependencies": {
				  		"grunt": "1.0.1",
				   		"@sap/grunt-sapui5-bestpractice-build": "^1.3.17"
				  }
				}
                '''
    }

    private emptyNameInJson() {
        return  '''
                {
				  "name": "",
				  "version": "1.0.3",
				  "description": "Webshop application for test purposes",
				  "private": true,
				  "devDependencies": {
				    "grunt": "1.0.1",
				    "@sap/grunt-sapui5-bestpractice-build": "^1.3.17"
				  }
				}
                '''
    }
    private missingVersionInJson() {
        return  '''
                {
				  "name": "com.mycompany.northwind",
				  "description": "Webshop application for test purposes",
				  "private": true,
				  "devDependencies": {
				    "grunt": "1.0.1",
				    "@sap/grunt-sapui5-bestpractice-build": "^1.3.17"
				  }
				}
                '''
    }

    private emptyVersionInJson() {
        return  '''
                {
				  	"name": "com.mycompany.northwind",
					"version": "",
				  	"description": "Webshop application for test purposes",
				  	"private": true,
				  	"devDependencies": {
				   		"grunt": "1.0.1",
				    		"@sap/grunt-sapui5-bestpractice-build": "^1.3.17"
				  	}
				}
                '''
    }
}
