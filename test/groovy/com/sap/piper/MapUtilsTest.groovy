package com.sap.piper

import org.junit.Assert
import org.junit.Test

class MapUtilsTest {

    @Test
    void testIsMap(){
        Assert.assertTrue('Map is not recognized as Map', MapUtils.isMap([:]))
        Assert.assertTrue('String is recognized as Map', !MapUtils.isMap('I am not a Map'))
    }
}
