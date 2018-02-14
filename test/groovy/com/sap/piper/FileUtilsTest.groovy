package com.sap.piper

import org.junit.Rule
import org.junit.Before
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.TemporaryFolder

import hudson.AbortException

import com.sap.piper.FileUtils


class FileUtilsTest {

    @Rule
    public ExpectedException thrown = new ExpectedException().none()

    @Rule
    public TemporaryFolder tmp = new TemporaryFolder()

    private File emptyDir
    private File notEmptyDir
    private File notDir


    @Before
    void setUp() {

        emptyDir = tmp.newFolder('emptyDir')
        notEmptyDir = tmp.newFolder('notEmptyDir')
        File file = new File("${notEmptyDir.getAbsolutePath()}${File.separator}test.txt")
        file.createNewFile()
        notDir = tmp.newFile('noDir')
    }

    @Test
    void nullValidateDirectoryTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'dir' can not be null or empty.")

        FileUtils.validateDirectory(null)
    }

    @Test
    void emptyValidateDirectoryTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'dir' can not be null or empty.")

        FileUtils.validateDirectory('')
    }

    @Test
    void doestNotExistValidateDirectoryTest() {

        def path = "${emptyDir.getAbsolutePath()}${File.separator}test"

        thrown.expect(AbortException)
        thrown.expectMessage("'$path' does not exist.")

        FileUtils.validateDirectory(path)
    }

    @Test
    void isNotDirectoryValidateDirectoryTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("'${notDir.getAbsolutePath()}' is not a directory.")

        FileUtils.validateDirectory(notDir.getAbsolutePath())
    }

    @Test
    void validateDirectoryTest() {

        FileUtils.validateDirectory(notEmptyDir.getAbsolutePath())
    }

    @Test
    void emptyDirValidateDirectoryIsNotEmptyTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("'${emptyDir.getAbsolutePath()}' is empty.")

        FileUtils.validateDirectoryIsNotEmpty(emptyDir.getAbsolutePath())
    }

    @Test
    void validateDirectoryIsNotEmptyTest() {

        FileUtils.validateDirectoryIsNotEmpty(notEmptyDir.getAbsolutePath())
    }
}
