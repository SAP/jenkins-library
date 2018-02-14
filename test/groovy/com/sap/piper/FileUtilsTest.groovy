package com.sap.piper

import org.junit.ClassRule
import org.junit.BeforeClass
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.TemporaryFolder

import hudson.AbortException

import com.sap.piper.FileUtils


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
}

