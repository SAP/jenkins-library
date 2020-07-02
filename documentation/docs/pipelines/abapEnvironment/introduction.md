# ABAP Environment Pipeline

The goal of the ABAP Environment Pipeline is to enable Continuous Integration for the SAP Cloud Platform ABAP Environment, also known as Steampunk.
In the current state, the pipeline enables you to pull Software Components to specifc systems and perform ATC checks. The following steps are performed:

* Create an instance of the SAP Cloud Platform ABAP Environment service
* Configure the Communication Arrangement SAP_COM_0510
* Pull Git repositories / Software Components to the instance
* Run ATC Checks
* Delete the SAP Cloud ABAP Environment system


