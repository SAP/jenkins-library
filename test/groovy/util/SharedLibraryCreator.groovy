package util

import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

class SharedLibraryCreator {

    static def lazyLoadedLibrary = getLibraryConfiguration(false)

    static def implicitLoadedLibrary = getLibraryConfiguration(true)

    private static def getLibraryConfiguration(def implicit) {
        library()
        .name('piper-library-os')
        .retriever(new ProjectSource())
        .targetPath('is/not/necessary')
        .defaultVersion("master")
        .allowOverride(true)
        .implicit(implicit)
        .build()
    }
}
