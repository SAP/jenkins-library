package com.sap.piper

import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat
import static org.junit.Assert.fail

import org.junit.Assert
import org.junit.Before
import org.junit.Test

import groovy.ui.SystemOutputInterceptor

class DefaultValueCacheTest {

    @Before
    void setup() {
        DefaultValueCache.reset()
    }

    @Test
    void testImmutable() {

        DefaultValueCache.createInstance(
            [myConfig:
                [
                    alphabet: (List)['a', new StringBuilder('b'), 'c', ['ae', 'oe', 'ue', 'sz'], [65: 'A', 66: 'B']],
                    bees: (Set)['carnica', 'buckfast'],
                    composers: [1685: 'J.S. Bach', 1882: 'Igor Strawinsky'],
                    missionStatment: new StringBuilder('Hello World!')]])

        immutabilityChecks(DefaultValueCache.getInstance())
    }

    @Test
    public void testImmutabilityAfterSerialization() {

        // This test here is important since with the CPS pattern serialization/deserialization happens frequently.

        DefaultValueCache.createInstance(
            [myConfig:
                [
                    alphabet: (List)['a', new StringBuilder('b'), 'c', ['ae', 'oe', 'ue', 'sz'], [65: 'A', 66: 'B']],
                    bees: (Set)['carnica', 'buckfast'],
                    composers: [1685: 'J.S. Bach', 1882: 'Igor Strawinsky'],
                    missionStatment: new StringBuilder().append('Hello World!')]])

        //
        // serialize and deserialize
        //
        // when working with ByteArrayOutputStream and - InputStream we do not allocate system resources, e.g.
        // Files. Everything is kept in Memory where normal garbage collection is in place. Hence we are a little
        // bit lazy here wrt closing the streams ...
        ByteArrayOutputStream byteOS = new ByteArrayOutputStream()
        ObjectOutputStream serialized = new ObjectOutputStream(byteOS)
        serialized.writeObject(DefaultValueCache.getInstance())
        serialized.flush()
        DefaultValueCache defaultValues = new ObjectInputStream(
            new ByteArrayInputStream(byteOS.toByteArray())).readObject()

        immutabilityChecks(defaultValues)
    }

    private static void immutabilityChecks(DefaultValueCache cache) {

        Map defaultValues = cache.getDefaultValues()

        try {
            defaultValues.myConfig.alphabet << '&'
            fail('We was able to add something to a unmodifiable entity.')
        } catch(UnsupportedOperationException e) {
        }

        try {
            defaultValues.myConfig.alphabet.each {
                it ->
                    if(it in List) it << '?'
            }
            fail('We was able to add something to a unmodifiable entity.')
        } catch(UnsupportedOperationException e) {
        }

        try {
            defaultValues.myConfig.alphabet.each {
                it ->
                    if(it in Map) it << [67: 'C']
            }
            fail('We was able to add something to a unmodifiable entity.')
        } catch(UnsupportedOperationException e) {
        }

        defaultValues.myConfig.alphabet.each {
            it ->

                if(it in CharSequence && it in StringBuilder)
                    fail "StringBuilder found: ${it}"
        }

        try {
            defaultValues.myConfig.bees << 'salmon'
            fail('We was able to add something to a unmodifiable entity.')
        } catch(UnsupportedOperationException e) {
        }

        try {
            defaultValues.myConfig.composers << [1888: "Friedrich Wilhelm Murnau"]
            fail('We was able to add something to a unmodifiable entity.')
        } catch(UnsupportedOperationException e) {
        }

        assert defaultValues.myConfig.missionStatment.getClass() in String

    }
}