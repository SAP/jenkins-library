# Docu stub

# Purpose
Here we should describe what this scenario does: It builds a SAPUI5 based application and deploys
the build result into a NEO account. Build is performed via mta -> node -> grunt -> sap-best-practices plugin
In fact it is `grunt clean, build, lint`.

# The following tools are used
 - mta (jar file which must be present on the build server; later: via docker)
 - npm (needs to be installed and configured on the build server: later: via docker)
 - grunt (will be materialized through npm)
 - sap best practices (will be materialized through npm)

# Prerequisites (in the central environment)

  - piper lib registered
  - npm installed and configured so that
      - Grunt and the
      - sap best practices build plugin can be materialized
  - Neo deploy account available, credentials maintained in Jenkins.

# The reader should already know

  - what a shared lib is and how it can be registered.
  - understanding of node /grunt build
  - SAP WEBIDE and the corresponding development workflow.

# Workflow
  - Developer works inside WEB-IDE and commits changes into this git clone.
  - Afterwards s?he pushes into a shared git repo.
  - Jenkins (push or pull?) detects the changes and triggers a build.
  - The build result gets deployed into the configured NEO account.

# References
- https://developers.sap.com/germany/tutorials/webide-grunt-basic.html

# Open questions:
  * What is our relationship to the SAP-WEB-IDE?

# Project template files

The following template files needs to be provided and adjusted on project level:

- *(.npmrc)[documentation/docs/scenarios/ui5-sap-cp/files/.npmrc]*
  Must contain a reference to the SAP npm registry: `@sap:registry https://npm.sap.com`.
  This dependency can be omitted on project lavel if it is provided in some higher configuration level (user, global).
  In this case it needs to be ensured that this dependency is available during local development as well as in central
  build infrastructure. Beside this the might might of course contain other npm configuration settings.
- *(mta.yaml)[documentation/docs/scenarios/ui5-sap-cp/files/mta.yaml]* Controls the behavior of the mta toolset. Placeholders
  (labeled like this `<placeholder`) needs to be replaced by valid values (version, applicationName). The `${timestamp}` in the version is replaced
  by the piperLibrary step `mtaBuild` in order to be able to distiguish the build results. <comment mh>I'm not happy with
  that approach modifying the sources here, we should check if this can be done in a better way.</comment mhend>
- *(package.json)[documentation/docs/scenarios/ui5-sap-cp/files/package.json]* Must contain a development dependency to `@sap/grunt-sapui5-bestpractice-build`. And of course other dependencies if required.
  Name, version and description needs to be maintained, too (placeholders).
- (Guntfile.js)[documentation/docs/scenarios/ui5-sap-cp/files/Gruntfile.js] controls the grunt build. By default these tasks are executed: `clean`, `build`, `lint`. It is
  less likely that this file needs to be changed.

# How we should extend our scenario
  - relationship to CM?
  - what about automated tests using the neo deploy space?
  - what about deployment into several neo accounts, e.g. for testing multiple aspects
  - what about deployment into production after successfully performed automated tests?
  - what about publishing lint results in order to create awareness about the lint results?
