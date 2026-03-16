/**
 * Shared utility functions for ABAP Environment Pipeline
 * These functions are reused across multiple pipeline stages to avoid code duplication
 */

/**
 * Checks if BTP (Business Technology Platform) mode is enabled based on configuration parameters.
 *
 * BTP mode is triggered when both subdomain and subaccount parameters are provided.
 * When BTP mode is enabled, the pipeline uses BTP CLI commands instead of Cloud Foundry commands.
 *
 * @param config Map containing pipeline configuration
 * @return boolean true if BTP mode is enabled, false otherwise
 *
 * Example usage:
 *   if (isBTPMode(config)) {
 *     btpCreateServiceInstance script: parameters.script
 *   } else {
 *     abapEnvironmentCreateSystem script: parameters.script
 *   }
 */
def isBTPMode(Object config) {
    // Ensure we have a map-like config (LinkedHashMap, Map, etc.)
    if (!(config instanceof Map)) {
        return false
    }

    // Check for both mandatory BTP parameters
    return config.subdomain && config.subaccount
}
