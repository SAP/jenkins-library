package com.sap.piper

enum DeploymentType {

    NEO_ROLLING_UPDATE('rolling-update'), CF_BLUE_GREEN('blue-green'), CF_STANDARD('standard'), NEO_DEPLOY('deploy')

    private String value

    public DeploymentType(String value){
        this.value = value
    }

    @Override
    public String toString(){
        return value
    }

    static DeploymentType selectFor(CloudPlatform cloudPlatform, boolean enableZeroDowntimeDeployment) {

        switch (cloudPlatform) {

            case CloudPlatform.NEO:
                if (enableZeroDowntimeDeployment) return NEO_ROLLING_UPDATE
                return NEO_DEPLOY

            case CloudPlatform.CLOUD_FOUNDRY:
                if (enableZeroDowntimeDeployment) return CF_BLUE_GREEN
                return CF_STANDARD

            default:
                throw new RuntimeException("Unknown cloud platform: ${cloudPlatform}")
        }
    }
}
