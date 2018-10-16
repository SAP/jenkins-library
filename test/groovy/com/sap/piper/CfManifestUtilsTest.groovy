package com.sap.piper

import org.junit.Test

import static org.junit.Assert.*

class CfManifestUtilsTest {

    @Test
    void testManifestTransform() {
        Map testFixture = [applications: [[buildpacks: ['sap_java_buildpack']]]]
        Map expected = [applications: [[buildpack: 'sap_java_buildpack']]]
        def actual = CfManifestUtils.transform(testFixture)
        assertEquals(expected, actual)
    }

    @Test(expected = RuntimeException)
    void testManifestTransformMultipleBuildpacks() {
        Map testFixture = [applications: [[buildpacks: ['sap_java_buildpack', 'another_buildpack']]]]
        CfManifestUtils.transform(testFixture)
    }

    @Test
    void testManifestTransformShouldNotChange() {
        Map testFixture = [applications: [[buildpack: 'sap_java_buildpack']]]
        Map expected = [applications: [[buildpack: 'sap_java_buildpack']]]
        def actual = CfManifestUtils.transform(testFixture)
        assertEquals(expected, actual)
    }
}
