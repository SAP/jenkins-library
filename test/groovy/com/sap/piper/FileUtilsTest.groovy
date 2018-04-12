package com.sap.piper

import org.junit.BeforeClass
import org.junit.ClassRule
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.TemporaryFolder
import org.junit.rules.RuleChain

import util.Rules

import com.lesfurets.jenkins.unit.BasePipelineTest

import com.sap.piper.FileUtils

import hudson.AbortException


class FileUtilsTest extends BasePipelineTest {

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)
                .around(thrown)

    private static emptyDir
    private static notEmptyDir
    private static file

    private static script

    @BeforeClass
    static void createTestFiles() {

        emptyDir = tmp.newFolder('emptyDir').getAbsolutePath()
        notEmptyDir = tmp.newFolder('notEmptyDir').getAbsolutePath()
        file = tmp.newFile('notEmptyDir/file.txt').getAbsolutePath()
    }

    @Before
    void setup() {

        script = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }


    @Test
    void validateDirectory_nullParameterTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'dir' can not be null or empty.")

        FileUtils.validateDirectory(script, null)
    }

    @Test
    void validateDirectory_emptyParameterTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'dir' can not be null or empty.")

        FileUtils.validateDirectory(script, '')
    }

    @Test
    void validateDirectory_directoryDoestNotExistTest() {

        def dir = new File("$emptyDir", 'test').getAbsolutePath()

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, dir) })

        thrown.expect(AbortException)
        thrown.expectMessage("Validation failed. '$dir' does not exist.")

        FileUtils.validateDirectory(script, dir)
    }

    @Test
    void validateDirectory_isNotDirectoryTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, file) })

        thrown.expect(AbortException)
        thrown.expectMessage("Validation failed. '$file' is not a directory.")

        FileUtils.validateDirectory(script, file)
    }

    @Test
    void validateDirectoryTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, notEmptyDir) })

        FileUtils.validateDirectory(script, notEmptyDir)
    }

    @Test
    void validateDirectoryIsNotEmpty_directoryIsEmptyTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, emptyDir) })

        thrown.expect(AbortException)
        thrown.expectMessage("Validation failed. '$emptyDir' is empty.")

        FileUtils.validateDirectoryIsNotEmpty(script, emptyDir)
    }

    @Test
    void validateDirectoryIsNotEmptyTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, notEmptyDir) })

        FileUtils.validateDirectoryIsNotEmpty(script, notEmptyDir)
    }

    @Test
    void validateFile_NoFilePathTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'filePath' can not be null or empty.")

        FileUtils.validateFile(script, null)
    }

    @Test
    void validateFile_emptyParameterTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'filePath' can not be null or empty.")

        FileUtils.validateFile(script, '')
    }

    @Test
    void validateFile_fileDoesNotExistTest() {

        def path = new File("$emptyDir", 'test').getAbsolutePath()

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, path) })

        thrown.expect(AbortException)
        thrown.expectMessage("Validation failed. '$path' does not exist.")

        FileUtils.validateFile(script, path)
    }

    @Test
    void validateFileTest() {

        helper.registerAllowedMethod('sh', [Map], { Map m -> script(m, file) })

        FileUtils.validateFile(script, file)
    }


    private script(parameters, path) {
      if(parameters.script.contains('exists')) return directoryOrFileExists(path)
      else if(parameters.script.contains('directory')) return isDirectory(path)
      else if(parameters.script.contains('empty')) return isDirectoryEmpty(path)
      else if(parameters.script.contains('file')) return isFile(path)
    }

    private directoryOrFileExists(dirOrFile) {
        def file = new File(dirOrFile)
        if (file.exists()) return 0
        else return 1
    }

    private isDirectory(dir) {
        def file = new File(dir)
        if (file.isDirectory()) return 0
        else return 1
    }

    private isDirectoryEmpty(dir) {
        def file = new File(dir)
        if (file.list().size() == 0) return 1
        return 0
    }

    private isFile(filePath) {
        def file = new File(filePath)
        if (file.isFile()) return 0
        return 1
    }
}
