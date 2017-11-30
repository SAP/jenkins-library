@Library('piper-library-os')

execute() {
    node() {
        setupCommonPipelineEnvironment script:this
    }
}

return this


