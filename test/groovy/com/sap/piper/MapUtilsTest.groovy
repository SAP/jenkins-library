package com.sap.piper

import org.junit.Assert
import org.junit.Test

class MapUtilsTest {

    @Test
    void testIsMap(){
        Assert.assertTrue('Map is not recognized as Map', MapUtils.isMap([:]))
        Assert.assertTrue('String is recognized as Map', !MapUtils.isMap('I am not a Map'))
    }

    @Test
    void testMergeMapStraigtForward(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2']],
             b = [b: '2',
                  c: [d: 'x']];

        Map merged = MapUtils.merge(a, b)

              assert merged == [a: '1',
                                b: '2',
                                c: [d: 'x', e: '2']]
    }

    @Test
    void testMergeMapNullValues(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2']],
             b = [b: '2',
                  c: null];

        Map merged = MapUtils.merge(a, b)

              assert merged == [a: '1',
                                b: '2',
                                c: [d: '1', e: '2']]
    }

}
