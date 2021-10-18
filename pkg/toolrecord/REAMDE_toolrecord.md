Purpose of the "toolrecord" feature is to provide a common result file for tools (e.g. code scanners) to allow consumers of the piper result files to locate the results in the respective tool backends for further reporting and post processing

Currently it contains the minimal information to detect which tools have been executed, and where to locate the results in the respective tool backends.

The result files are called "tr_toolname_YYYYMMDDHHMMSS.json" and have the following structure:

{
    "RecordVersion":1,
    "ToolName":"dummyTool",
    "ToolInstance":"dummyInstance",   // Tool backend URL

    // Tool-agnostic DisplayName and DisplayUrl for simple reportings
    // ( this is deried from the keys details )
    "DisplayName":"dummyOrgName - dummyProjName - dummyScanName",
    "DisplayURL":"dummyScanUrl",

    // tool-dependend identifiers; order is taken of tool's data model e.g. 'team owns project has scan'
    "Keys":[
        {
            "Name":"Organization",         // the technical name from the tool's data model
            "Value":"dummyOrgId",          // the key value needed to access the tool's backend via api
            "DisplayName":"dummyOrgName",  // User-friendly identifiert - optional can be empty
            "URL":"dummyOrgUrl"            // Url to access this data in the tool's ui - optional can be empty
        },
        {"Name":"Project","Value":"dummyProjectId","DisplayName":"dummyProjName","Url":"dummyProjUrl"},
        {"Name":"ScanId","Value":"dummyScanId","DisplayName":"dummyScanName","Url":"dummyScanUrl"}
        ],

    "Context":{}                            // additional context data - optional tool dependend
}
