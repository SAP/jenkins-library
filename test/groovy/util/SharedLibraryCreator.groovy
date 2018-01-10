package util

import static com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration.library

class SharedLibraryCreator {

    static def lazyLoadedLibrary = library()
        .name('piper-library')
        .retriever(new ProjectSource())
        .targetPath('is/not/necessary')
        .defaultVersion("master")
        .allowOverride(true)
        .implicit(false)
        .build()

    static def implicitLoadedLibrary = library()
        .name('piper-library')
        .retriever(new ProjectSource())
        .targetPath('is/not/necessary')
        .defaultVersion("master")
        .allowOverride(true)
        .implicit(true)
        .build()
}
