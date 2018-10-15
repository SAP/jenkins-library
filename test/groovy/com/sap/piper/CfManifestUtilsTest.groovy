package com.sap.piper

import org.junit.Assert
import org.junit.Test

class CfManifestUtilsTest {
    Map testFixture = [applications: [[buildpacks: ['sap_java_buildpack']]]]

    @Test
    void testManifestTransform() {
        Map expected = [applications: [[buildpack: 'sap_java_buildpack']]]
        def actual = CfManifestUtils.transform(testFixture)
        Assert.assertEquals(expected, actual)
    }
}
