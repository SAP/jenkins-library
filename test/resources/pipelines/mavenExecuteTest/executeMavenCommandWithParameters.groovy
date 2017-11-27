@Library('piper-library-os')

execute() {
    node() {
        mavenExecute(
            script: this,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            goals: 'clean install',
            globalSettingsFile: 'globalSettingsFile.xml',
            projectSettingsFile: 'projectSettingsFile.xml',
            pomPath: 'pom.xml',
            flags: '-o',
            m2Path: 'm2Path',
            defines: '-Dmaven.tests.skip=true'
        )
    }
}

return this


