package com.sap.piper

import org.junit.Assert
import org.junit.Test

import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class MapUtilsTest {

    @Test
    void testIsMap(){
        Assert.assertTrue('Map is not recognized as Map', MapUtils.isMap([:]))
        Assert.assertTrue('String is recognized as Map', !MapUtils.isMap('I am not a Map'))
        Assert.assertFalse('Null value is recognized as Map', MapUtils.isMap(null))
    }

    @Test
    void testMergeMapStraightForward(){

        Map a = [a: '1',
                 c: [d: '1',
                     e: '2']],
             b = [b: '2',
                  c: [d: 'x']]

        Map merged = MapUtils.merge(a, b)

        assert merged == [a: '1',
                          b: '2',
                          c: [d: 'x', e: '2']]
    }

    @Test
    void testMergeMapWithConflict(){

        Map a = [a: '1',
                 b: [c: 1]],
            b = [a: '2',
                 b: [c: 2]]

        Map merged = MapUtils.merge(a, b)

        assert merged == [a: '2',
                          b: [c: 2]]
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

    @Test
    void testGetByPath() {
        Map m = [trees: [oak: 5, beech :1], flowers:[rose: 23]]

        assertThat(MapUtils.getByPath(m, 'flowers'), is([rose: 23]))
        assertThat(MapUtils.getByPath(m, 'trees/oak'), is(5))
        assertThat(MapUtils.getByPath(m, 'trees/palm'), is(null))
    }

    @Test
    void testDeepCopy() {

        List l = ['a', 'b', 'c']

        def original = [
                list: l,
                set: (Set)['1', '2'],
                nextLevel: [
                    list: ['x', 'y'],
                    duplicate: l,
                    set: (Set)[9, 8, 7]
                ]
            ]

        def copy = MapUtils.deepCopy(original)

        assert ! copy.is(original)
        assert ! copy.list.is(original.list)
        assert ! copy.set.is(original.set)
        assert ! copy.nextLevel.list.is(original.nextLevel.list)
        assert ! copy.nextLevel.set.is(original.nextLevel.set)
        assert ! copy.nextLevel.duplicate.is(original.nextLevel.duplicate)

        // Within the original identical list is used twice, but the
        // assuption is that there are different lists in the copy.
        assert ! copy.nextLevel.duplicate.is(copy.list)

        assert copy == original
    }
}
