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
    void testMergeMapNullValueDoesNotOverwriteNonNullValue(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2']],
             b = [b: '2',
                  c: null];

        Map merged = MapUtils.merge(a, b)

        assert merged == [a: '1',
                          b: '2',
                          c: [d: '1', e: '2']] // <-- here we do not have null, since skipNull defaults to true
    }

    @Test
    void testMergeMapNullValueOverwritesNonNullValueWhenSkipNullIsFalse(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2']],
             b = [b: '2',
                  c: null];

        Map merged = MapUtils.merge(a, b, false)

        assert merged == [a: '1',
                          b: '2',
                          c: null] // <-- here we have null, since we have skipNull=false
    }
    @Test
    void testMergeMapNullNullValueIsPreservedFromOverlayMapIfNotInBaseMap(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2']],
             b = [b: '2',
                  c: null, // <-- Will not be taken into account, but the entry from the base map will be present.
                  n: null];// <-- Will not be taken into account.

        Map merged = MapUtils.merge(a, b)

        assert merged == [a: '1',
                          b: '2',
                          c: [d: '1', e: '2']]
    }

    @Test
    void testMergeMapNullValueInBaseMapIsPreserved(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2'],
                 n: null], // <-- This entry will be preserved.
             b = [b: '2',
                  c: [d: 'x']];

        Map merged = MapUtils.merge(a, b)

        assert merged == [a: '1',
                          b: '2',
                          n: null,
                          c: [d: 'x', e: '2']]
    }

}
