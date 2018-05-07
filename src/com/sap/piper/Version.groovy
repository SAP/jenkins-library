package com.sap.piper

import hudson.AbortException


class Version implements Serializable {

    final def major
    final def minor
    final def patch

    Version(major, minor, patch = -1) {
      if (major < 0) throw new IllegalArgumentException("The parameter 'major' can not have a value less than 0.")
      if (minor < 0) throw new IllegalArgumentException("The parameter 'minor' can not have a value less than 0.")
      this.major = major
      this.minor = minor
      this.patch = patch
    }

    Version(text) {
        if (!text) throw new IllegalArgumentException("The parameter 'text' can not be null or empty.")
        def group = text =~ /(\d+[.]\d+[.]\d+)/
        if (!group) throw new AbortException("The version '$text' has an unexpected format. The expected format is <major.minor.patch>.")
        def i = group[0].size()-1
        def versionNumbers = group[0][i].split("\\.")
        major = versionNumbers[0].toInteger()
        minor = versionNumbers[1].toInteger()
        patch = versionNumbers[2].toInteger()
    }

    @Override
    boolean equals(version) {
        if (!version) throw new IllegalArgumentException("The parameter 'version' can not be null.")
        return major == version.major && minor == version.minor && patch == version.patch
    }

    boolean isHigher(version) {
        if (!version) throw new IllegalArgumentException("The parameter 'version' can not be null.")
        return major > version.major || major == version.major && ( minor > version.minor || minor == version.minor && patch > version.patch)
    }

    boolean isCompatibleVersion(version) {
        if (!version) throw new IllegalArgumentException("The parameter 'version' can not be null.")
        return this == version || isHigher(version) && major == version.major
    }

    @Override
    String toString() {
        return patch != -1 ? "$major.$minor.$patch".toString() : "$major.$minor".toString()
    }
}

