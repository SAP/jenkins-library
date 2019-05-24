package com.sap.piper.k8s

import com.sap.piper.k8s.SystemEnv
import org.junit.Before
import org.junit.Test

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class SystemEnvTest {
    SystemEnv env = null
    Map systemEnvironmentMock = [:]
    @Before
    void setUp() {
        systemEnvironmentMock = ['HTTP_PROXY' : 'http://my-http-proxy:8080',
                                 'HTTPS_PROXY': 'http://my-http-proxy:8080',
                                 'NO_PROXY'   : '*.example.com,localhost',
                                 'http_proxy' : 'http://my-http-proxy:8080',
                                 'https_proxy': 'http://my-http-proxy:8080',
                                 'no_proxy'   : '*.example.com,localhost',]
        System.metaClass.static.getenv = { String s -> return systemEnvironmentMock.get(s) }
        env = new SystemEnv()
    }

    @Test
    void testget() {
        String name = 'HTTP_PROXY'
        assertEquals(systemEnvironmentMock.get(name), env.get(name))

        name = 'HTTPS_PROXY'
        assertEquals(systemEnvironmentMock.get(name), env.get(name))

    }

    @Test
    void testgetEnv() {
        assertNotNull(env)
        assertEquals(systemEnvironmentMock.keySet(), env.getEnv().keySet())
    }

    @Test
    void testremove() {
        String name = 'HTTP_PROXY'
        env.remove(name)
        assertEquals(env.getEnv().containsKey(name),false)
    }
}
