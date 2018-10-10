package com.sap.piper

import org.junit.Assert
import org.junit.Test

class MapUtilsTest {

    @Test
    void testIsMap(){
        Assert.assertTrue('Map is not recognized as Map', MapUtils.isMap([:]))
        Assert.assertTrue('String is recognized as Map', !MapUtils.isMap('I am not a Map'))
        Assert.assertFalse('Null value is recognized as Map', MapUtils.isMap(null))
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
    void testPruneNulls() {

        Map m = [a: '1',
                 b: 2,
                 c: [ d: 'abc',
                      e: '',
                      n2: null],
                 n1: null]

        assert MapUtils.pruneNulls(m) == [ a: '1',
                                           b: 2,
                                           c: [ d: 'abc',
                                                e: '']]
    }

    @Test
    void testTraverse() {
        Map m = [a: 'x1', m:[b: 'x2', c: 'otherString']]
        MapUtils.traverse(m, { s -> (s.startsWith('x')) ? "replaced" : s})
        assert m == [a: 'replaced', m: [b: 'replaced', c: 'otherString']]
    }
}
