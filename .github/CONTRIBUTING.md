# Guidance on How to Contribute

**Table of Contents:**

1. [Using the issue tracker](#using-the-issue-tracker)
1. [Changing the code-base](#changing-the-code-base)
1. [Jenkins credential handling](#jenkins-credential-handling)
1. [Code Style](#code-style)
1. [References](#references)

There are two primary ways to help:

* Using the issue tracker, and
* Changing the code-base.

## Using the issue tracker

Use the issue tracker to suggest feature requests, report bugs, and ask questions. This is also a great way to connect with the developers of the project as well as others who are interested in this solution.

Use the issue tracker to find ways to contribute. Find a bug or a feature, mention in the issue that you will take on that effort, then follow the "Changing the code-base" guidance below.

## Changing the code-base

Generally speaking, you should fork this repository, make changes in your own fork, and then submit a pull-request. All new code should have been thoroughly tested end-to-end in order to validate implemented features and the presence or lack of defects.

### Working with forks

* [Configure this repository as a remote for your own fork](https://help.github.com/articles/configuring-a-remote-for-a-fork/), and
* [Sync your fork with this repository](https://help.github.com/articles/syncing-a-fork/) before beginning to work on a new pull-request.

### Tests

All pipeline library coding _must_ come with automated unit tests.

Besides that, we have an integration test suite, which is not triggered during normal pull request builds. However, integration tests are mandatory before a change can be merged. It is the duty of a team member of the SAP/jenkins-library project to execute these tests.
To trigger the integration test suite, the `HEAD` commit of the branch associated with the pull request must be pushed under the branch pattern `it/.*` (recommended naming convention: `it/<Number of the pull request>`). As a result, the status `integration-tests` is updated in the pull request.

### Documentation

The contract of functionality exposed by a library functionality needs to be documented, so it can be properly used.
Implementation of a functionality and its documentation shall happen within the same commit(s).

### Coding pattern

Pipeline steps must not make use of return values. The pattern for sharing parameters between pipeline steps or between a pipeline step and a pipeline script is sharing values via the [`commonPipelineEnvironment`](../vars/commonPipelineEnvironment.groovy). Since there is no return value from a pipeline step the return value of a pipeline step is already `void` rather than `def`.

#### EditorConfig

To ensure a common file format, there is a `.editorConfig` file [in place](../.editorconfig). To respect this file, [check](http://editorconfig.org/#download) if your editor does support it natively or you need to download a plugin.

### Commit Message Style

Write [meaningful commit messages](http://who-t.blogspot.de/2009/12/on-commit-messages.html) and [adhere to standard formatting](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

Good commit messages speed up the review process and help to keep this project maintainable in the long term.

## Developer Certificate of Origin (DCO)

Due to legal reasons, contributors will be asked to accept a DCO when they create their first pull request to this project. This happens in an automated fashion during the submission process. SAP uses [the standard DCO text of the Linux Foundation](https://developercertificate.org/).

## Jenkins credential handling

References to Jenkins credentials should have meaningful names.

We are using the following approach for naming Jenkins credentials:

For username/password credentials:
`<tool>CredentialsId` like e.g. `neoCredentialsId`

For other cases we add further information to the name like:

* `gitSshCredentialsId` for ssh credentials
* `githubTokenCredentialsId`for token/string credentials
* `gcpFileCredentialsId` for file credentials

## Code Style

Generally, the code should follow any stylistic and architectural guidelines prescribed by the project. In the absence of guidelines, mimic the styles and patterns in the existing code-base.

The intention of this section is to describe the code style for this project. As reference document, the [Groovy's style guide](http://groovy-lang.org/style-guide.html) was taken. For further reading about Groovy's syntax and examples, please refer to this guide.

This project is intended to run in Jenkins [[2]](https://jenkins.io/doc/book/getting-started/) as part of a Jenkins Pipeline [[3]](https://jenkins.io/doc/book/pipeline/). It is composed by Jenkins Pipeline's syntax, Groovy's syntax and Java's syntax.

Some Groovy's syntax is not yet supported by Jenkins. It is also the intention of this section to remark which Groovy's syntax is not yet supported by Jenkins.

As Groovy supports 99% of Java’s syntax [[1]](http://groovy-lang.org/style-guide.html), many Java developers tend to write Groovy code using Java's syntax. Such a developer should also consider the following code style for this project.

### General remarks

Variables, methods, types and so on shall have meaningful self describing names. Doing so makes understanding code easier and requires less commenting. It helps people who did not write the code to understand it better.

Code shall contain comments to explain the intention of the code when it is unclear what the intention of the author was. In such cases, comments should describe the "why" and not the "what" (that is in the code already).

### Omit semicolons

### Use the return keyword

In Groovy it is optional to use the _return_ keyword. Use explicitly the _return_ keyword for better readability.

### Use def

When using _def_ in Groovy, the type is Object. Using _def_ simplifies the code, for example imports are not needed, and therefore the development is faster.

### Do not use a visibility modifier for public classes and methods

By default, classes and methods are public, the use of the public modifier is not needed.

### Do not omit parentheses for Groovy methods

In Groovy is possible to omit parentheses for top-level expressions, but [Jenkins Pipeline's syntax](https://jenkins.io/doc/book/pipeline/syntax/) use a block, specifically `pipeline { }` as top-level expression [[4]](https://jenkins.io/doc/book/pipeline/syntax/). Do not omit parenthesis for Groovy methods because Jenkins will interpret the method as a Pipeline Step. Conversely, do omit parenthesis for Jenkins Pipeline's Steps.

### Omit the .class suffix

In Groovy, the .class suffix is not needed. Omit the .class suffix for simplicity and better readability.

e.g. `new ExpectedException().expect(AbortException.class)`

-->  `new ExpectedException().expect(AbortException)`

### Omit getters and setters

When declaring a field without modifier inside a Groovy bean, the Groovy compiler generates a private field and a getter and setter.

### Do not initialize beans with named parameters

Do not initialize beans with named parameters, because it is not supported by Jenkins:

e.g. `Version javaVersion = new Version( major: 1, minor: 8)`

Initialize beans using Java syntax:

e.g. `Version javaVersion = new Version(1, 8)`

Use named parameters for Jenkins Pipeline Steps:

e.g. `sh returnStdout: true, script: command`

### Do not use _with()_ operator

The _with_ operator is not yet supported by Jenkins, and it must not be used or encapsulated in a @NonCPS method.

### Use _==_ operator

Use Groovy’s `==` instead of Java `equals()` to avoid NullPointerExceptions. To compare the references of objects, instead of `==`, you should use `a.is(b)` [[1]](http://groovy-lang.org/style-guide.html).

### Use GStrings

In Groovy, single quotes create Java Strings, and double quotes can create Java Strings or GStrings, depending if there is or not interpolation of variables [[1]](http://groovy-lang.org/style-guide.html). Using GStrings variable and string concatenation is more simple.

#### Do not use curly braces {} for variables or variable.property

For variables, or variable.property, drop the curly braces:

e.g. `echo "[INFO] ${name} version ${version.version} is installed."`

-->  `echo "[INFO] $name version $version.version is installed."`

#### Use 'single quotes' for Strings and constants

#### Use "double quotes" for GStrings

#### Use '''triple single quotes''' for multiline Strings

#### Use """triple double quotes""" for multiline GStrings

#### Use /slash/ for regular expressions

This notation avoids to double escape backslashes, making easier working with regex.

### Use native syntax for data structures

Use the native syntax for data structures provided by Groovy like lists, maps, regex, or ranges of values.

### Use aditional Groovy methods

Use the additional methods provided by Groovy to manipulate String, Files, Streams, Collections, and other classes.
For a complete description of all available methods, please read the GDK API [[5]](http://groovy-lang.org/groovy-dev-kit.html).

### Use Groovy's switch

Groovy’s switch accepts any kind of type, thereby is more powerful. In this case, the use of _def_ instead of a type is necessary.

### Use alias for import

In Groovy, it is possible to assign an alias to imported packages. Use alias for imported packages to avoid the use of fully-qualified names and increase readability.

### Use Groovy syntax to check objects

In Groovy a null, void, equal to zero, or empty object evaluates to false, and if not, evaluates to true. Instead of writing null and size checks e.g. `if (name != null && name.length > 0) {}`, use just the object `if (name) {}`.

### Use _?._ operator

Use the safe dereference operator  _?._, to simplify the code for accessing objects and object members safely. Using this operator, the Groovy compiler checks null objects and null object members, and returns _null_ if the object or the object member is null and never throws a NullPointerException.

### Use _?:_ operator

Use Elvis operator _?:_ to simplify default value validations.

### Use _any_ keyword

If the type of the exception thrown inside a try block is not important, catch any exception using the _any_ keyword.

### Use _assert_

To check parameters, return values, and more, use the assert statement.

## References

[1] Groovy's syntax: [http://groovy-lang.org/style-guide.html](http://groovy-lang.org/style-guide.html)

[2] Jenkins: [https://jenkins.io/doc/book/getting-started/](https://jenkins.io/doc/book/getting-started/)

[3] Jenkins Pipeline: [https://jenkins.io/doc/book/pipeline/](https://jenkins.io/doc/book/pipeline/)

[4] Jenkins Pipeline's syntax: [https://jenkins.io/doc/book/pipeline/syntax/](https://jenkins.io/doc/book/pipeline/syntax/)

[5] GDK: Groovy Development Kit: [http://groovy-lang.org/groovy-dev-kit.html](http://groovy-lang.org/groovy-dev-kit.html)
