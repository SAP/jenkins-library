package com.sap.piper

import org.junit.Before
import org.junit.Test

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class ContainerMapTest {
    @Before
    void setUp() {

    }

    @Test
    void testIfObjectCreated() {
        assertNotNull(ContainerMap.instance)
    }

    @Test
    void testSetMap() {
        ContainerMap.instance.setMap(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']])
        assertEquals(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']],ContainerMap.instance.getMap())
    }

    @Test
    void testGetMap() {
        assertNotNull(ContainerMap.instance.getMap())
    }
}
