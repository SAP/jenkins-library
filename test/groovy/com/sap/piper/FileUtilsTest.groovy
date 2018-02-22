package com.sap.piper

import org.junit.*
import org.junit.rules.*

import com.sap.piper.FileUtils
import hudson.AbortException


class FileUtilsTest {

    @ClassRule
    public static TemporaryFolder tmp = new TemporaryFolder()

    @Rule
    public ExpectedException thrown = new ExpectedException()

    private static emptyDir
    private static notEmptyDir
    private static file

    @BeforeClass
    static void createTestFiles() {

        emptyDir = tmp.newFolder('emptyDir').getAbsolutePath()
        notEmptyDir = tmp.newFolder('notEmptyDir').getAbsolutePath()
        file = tmp.newFile('notEmptyDir/file').getAbsolutePath()
    }


    @Test
    void nullValidateDirectoryTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'dir' can not be null or empty.")

        FileUtils.validateDirectory()
    }

    @Test
    void emptyValidateDirectoryTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'dir' can not be null or empty.")

        FileUtils.validateDirectory('')
    }

    @Test
    void doestNotExistValidateDirectoryTest() {

        def path = new File("$emptyDir", 'test').getAbsolutePath()

        thrown.expect(AbortException)
        thrown.expectMessage("'$path' does not exist.")

        FileUtils.validateDirectory(path)
    }

    @Test
    void isNotDirectoryValidateDirectoryTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("'$file' is not a directory.")

        FileUtils.validateDirectory(file)
    }

    @Test
    void validateDirectoryTest() {

        FileUtils.validateDirectory(notEmptyDir)
    }

    @Test
    void emptyDirValidateDirectoryIsNotEmptyTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("'$emptyDir' is empty.")

        FileUtils.validateDirectoryIsNotEmpty(emptyDir)
    }

    @Test
    void validateDirectoryIsNotEmptyTest() {

        FileUtils.validateDirectoryIsNotEmpty(notEmptyDir)
    }

    @Test
    void validateFileNoFilePathTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'filePath' can not be null or empty.")

        FileUtils.validateFile(null)
    }

    @Test
    void validateFileEmptyFilePathTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'filePath' can not be null or empty.")

        FileUtils.validateFile('')
    }

    @Test
    void validateFileDoesNotExistFileTest() {

        def path = new File("$emptyDir", 'test').getAbsolutePath()

        thrown.expect(AbortException)
        thrown.expectMessage("'$path' does not exist.")

        FileUtils.validateFile(path)
    }

    @Test
    void validateFileTest() {

        FileUtils.validateFile(file)
    }
}

