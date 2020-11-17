# Continuous Testing on SAP Cloud Platform ABAP Environment

## Introduction

This scenario describes how to test ABAP development for the SAP Cloud Platform ABAP Environment (also known as Steampunk). In Steampunk, the development is done within [“software components”](https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/58480f43e0b64de782196922bc5f1ca0.html) (also called: “repositories”) and "transported" via git-based approaches. The [ABAP Environment Pipeline](../pipelines/abapEnvironment/introduction.md) is a predefined pipeline, which can be used to import ABAP development into a quality system and execute tests.

## Pipeline

For this scenario three stages of the ABAP Environment Pipeline are relevant: "Prepare System", "Clone Repositories" and "ATC".

### Prepare System

The pipeline starts with the stage "Prepare System". This stage, however, is optional.  **If this stage is active**, a new Steampunk system is created for each pipeline execution. This has the advantage, that each test runs on a fresh system without a history. On the other hand, the duration of each pipeline execution will increase as the system provisioning takes a significant amount of time. **If this stage is not active**, you have to provide a prepared Steampunk (quality) system for the other stages. Then, each pipeline execution runs on the same system. Of course, the system has a history, but the pipeline durtion will be shorter. Please also consider: the total costs may increase for a static system in contrast to a system, which is only active during the pipeline.

### Clone Repositories

This stage is responsible for cloning (or pulling) the defined software components (repositories) to the system.
