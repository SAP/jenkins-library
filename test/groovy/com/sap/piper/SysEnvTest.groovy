package com.sap.piper

import org.junit.Before
import org.junit.Test

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class SysEnvTest {
    SysEnv env = null;

    @Before
    void setUp() throws Exception {
        env = new SysEnv()
        assertNotNull(env)
    }

    @Test
    void testget() {
        String name = 'HTTP_PROXY'
        assertEquals(env.get(),System.getenv(name))

        name = 'HTTPS_PROXY'
        assertEquals(env.get(),System.getenv(name))

    }

    @Test
    void testgetEnv() {
        Map envVars = env.getEnv()
        String name = 'HTTP_PROXY'
        assertNotNull(envVars)
        assertEquals(envVars.get(name),env.get(name))
    }

    @Test
    void testremove() {
        String name = 'HTTP_PROXY'
        env.remove(name)
        assertEquals(env.getEnv().containsKey(name),false)
    }

}
