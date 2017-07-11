package com.sap.piper

import hudson.AbortException
import java.io.File


class FileUtils implements Serializable {

    static validateDirectory(dir) {
        if (!dir) throw new IllegalArgumentException("The parameter 'dir' can not be null or empty.")
        def file = new File(dir)
        if (!file.exists()) throw new AbortException("'${file.getAbsolutePath()}' does not exist.")
        if (!file.isDirectory()) throw new AbortException("'${file.getAbsolutePath()}' is not a directory.")
    }

    static validateDirectoryIsNotEmpty(dir) {
        validateDirectory(dir)
        def file = new File(dir)
        if (file.list().size() == 0) throw new AbortException("'${file.getAbsolutePath()}' is empty.")
    }
}
