# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

!!! note "Breaking change in `goals`, `defines` and `flags` parameters"
    The `goals`, `defines` and `flags` parameters of the step need to be lists of strings with each element being one item.

    As an example consider this diff, showing the old api deleted and the new api inserted:

    ```diff
    -goals: 'org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate',
    -defines: "-Dexpression=$pomPathExpression -DforceStdout -q",
    +goals: ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
    +defines: ["-Dexpression=$pomPathExpression", "-DforceStdout", "-q"],
    ```

    Additionally please note that in the parameters _must not_ be [shell quoted/escaped](https://www.tldp.org/LDP/Bash-Beginners-Guide/html/sect_03_03.html).
    What you pass in is literally passed to Maven without any shell interpreter in between.

    The old behavior is still available in version `v1.23.0` and before of project "Piper".

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Exceptions

None

## Example

```groovy
mavenExecute script: this, goals: ['clean', 'install']
```

Example for the correct usage of `goals`, `defines` and `flags` in version `v1.24.0` and newer:

```groovy
mavenExecute(
    script: script,
    goals: ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
    defines: ["-Dexpression=$pomPathExpression", "-DforceStdout", "-q"],
    returnStdout: true
)
```

Note that it does not work to put multiple arguments into one element of a list, so `defines: ["-Dexpression=$pomPathExpression -DforceStdout -q"]` does **not** work.
