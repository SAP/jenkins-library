@Library('piper-library-os')

execute() {
    node() {
        mavenExecute script: this, goals: 'clean install'
    }
}

return this
