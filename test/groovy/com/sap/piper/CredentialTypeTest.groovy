package com.sap.piper

import org.junit.Assert
import org.junit.Test

class CredentialTypeTest {

    @Test
    void "Check that enum literals have not changed"() {
        assert "${CredentialType.FILE}" == 'file'
        assert "${CredentialType.TOKEN}" == 'token'
        assert "${CredentialType.SECRET_TEXT}" == 'secretText'
        assert "${CredentialType.USERNAME_PASSWORD}" == 'usernamePassword'
        assert "${CredentialType.SSH}" == 'ssh'
    }

}
