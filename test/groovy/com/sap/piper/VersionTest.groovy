package com.sap.piper

import org.junit.Rule
import org.junit.Before
import org.junit.Test
import org.junit.rules.ExpectedException

import hudson.AbortException

import com.sap.piper.Version


class VersionTest {


    @Rule
    public ExpectedException thrown = new ExpectedException().none()


    @Test
    void illegalMajorVersionTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'major' can not have a value less than 0.")

        Version version = new Version(-1,0)
    }

    @Test
    void illegalMinorVersionTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'minor' can not have a value less than 0.")

        Version version = new Version(0,-1)
    }

    @Test
    void nullMajorVersionTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'major' can not have a value less than 0.")

        Version version = new Version(null,0)
    }

    @Test
    void nullMinorVersionTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'minor' can not have a value less than 0.")

        Version version = new Version(0, null)
    }

    @Test
    void nullVersionTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'text' can not be null or empty.")

        Version version = new Version(null)
    }

    @Test
    void emptyVersionTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'text' can not be null or empty.")

        Version version = new Version('')
    }

    @Test
    void unexpectedFormatTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("The version '0-0.1' has an unexpected format. The expected format is <major.minor.patch>.")

        Version version = new Version('0-0.1')
    }

    @Test
    void isEqualNullTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'version' can not be null.")

        Version version = new Version(0,0,1)
        version.equals(null)
    }

    @Test
    void isEqualPatchTest() {

        Version version1 = new Version(0,0,1)
        Version version2 = new Version('0.0.1')

        assert version1.equals(version2)
    }

    @Test
    void isEqualMinorTest() {

        Version version1 = new Version(0,1,0)
        Version version2 = new Version('0.1.0')

        assert version1.equals(version2)
    }

    @Test
    void isEqualMajorTest() {

        Version version1 = new Version(1,0,0)
        Version version2 = new Version('1.0.0')

        assert version1.equals(version2)
    }

    @Test
    void isHigherNullTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'version' can not be null.")

        Version version = new Version(0,0,1)
        version.isHigher(null)
    }

    @Test
    void isHigherPatchTest() {

        Version version1 = new Version(0,0,1)
        Version version2 = new Version('0.0.2')

        assert version2.isHigher(version1)
    }

    @Test
    void isHigherMinorTest() {

        Version version1 = new Version(0,1,0)
        Version version2 = new Version('0.2.0')

        assert version2.isHigher(version1)
    }

    @Test
    void isHigherMajorTest() {

        Version version1 = new Version(1,0,0)
        Version version2 = new Version('2.0.0')

        assert version2.isHigher(version1)
    }

    @Test
    void isCompatibleVersionNullTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("The parameter 'version' can not be null.")

        Version version = new Version(0,0,1)
        version.isCompatibleVersion(null)
    }

    @Test
    void isCompatibleVersionPatchTest() {

        Version version1 = new Version(0,0,1)
        Version version2 = new Version('0.0.2')

        assert version2.isCompatibleVersion(version1)
    }

    @Test
    void isCompatibleVersionMinorTest() {

        Version version1 = new Version(0,1,0)
        Version version2 = new Version('0.2.0')

        assert version2.isCompatibleVersion(version1)
    }

    @Test
    void isIncompatibleVersionTest() {

        Version version1 = new Version(1,0,0)
        Version version2 = new Version('2.0.0')

        assert !version2.isCompatibleVersion(version1)
    }
}

