package com.sap.piper

import hudson.AbortException


class EnvironmentUtils implements Serializable {

    static boolean isEnvironmentVariable(script, variable) {
        return !getEnvironmentVariable(script, variable).isEmpty()
    }

    static String getEnvironmentVariable(script, variable) {
        try {
          def envVar = script.sh returnStdout: true, script: """#!/bin/bash --login
                                                                echo \$$variable"""
          return envVar.trim()
        } catch(AbortException e) {
          throw new AbortException("There was an error requesting the environment variable '$variable'. Reason: $e.message.")
        }
    }
}
