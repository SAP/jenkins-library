package com.sap.piper.versioning

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals

class ArtifactVersioningTest {

    @Rule
    public ExpectedException thrown = ExpectedException.none()

    @Test
    void testInstatiateFactoryMethod() {
        def versionObj = ArtifactVersioning.getArtifactVersioning( 'maven', this, [:])
        assertTrue(versionObj instanceof MavenArtifactVersioning)
    }

    @Test
    void testInstatiateFactoryMethodWithInvalidToolId() {
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage('No versioning implementation for buildTool: invalid available.')
        ArtifactVersioning.getArtifactVersioning('invalid', this, [:])
    }
}
