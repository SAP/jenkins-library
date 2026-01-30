# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

### Setting up InfluxDB with Grafana

The easiest way to start with is using the available official docker images.
You can either run these docker containers on the same host on which you run your Jenkins or each docker on individual VMs (hosts).
Very basic setup can be done like that (with user "admin" and password "adminPwd" for both InfluxDB and Grafana):

    docker run -d -p 8083:8083 -p 8086:8086 --restart=always --name influxdb -v /var/influx_data:/var/lib/influxdb influxdb
    docker run -d -p 3000:3000 --name grafana --restart=always --link influxdb:influxdb -e "GF_SECURITY_ADMIN_PASSWORD=adminPwd" grafana/grafana

For more advanced setup please reach out to the respective documentation:

- InfluxDB ([Docker Hub](https://hub.docker.com/_/influxdb/) [GitHub](https://github.com/docker-library/docs/tree/master/influxdb))
- Grafana ([Docker Hub](https://hub.docker.com/r/grafana/grafana/) [GitHub](https://github.com/grafana/grafana-docker))

After you have started your InfluxDB docker you need to create a database:

- in a Webbrowser open the InfluxDB Web-UI using the following URL: &lt;host of your docker&gt;:8083 (port 8083 is used for access via Web-UI, for Jenkins you use port 8086 to access the DB)
- create new DB (the name of this DB you need to provide later to Jenkins)
- create Admin user (this user you need to provide later to Jenkins)

!!! hint "With InfluxDB version 1.1 the InfluxDB Web-UI is deprecated"

You can perform the above steps via commandline:

- The following command will create a database with name &lt;databasename&gt;

  `curl -i -XPOST http://localhost:8086/query --data-urlencode "q=CREATE DATABASE \<databasename\>"`

- The admin user with the name &lt;adminusername&gt; and the password &lt;adminuserpwd&gt; can be created with

  `curl -i -XPOST http://localhost:8086/query --data-urlencode "q=CREATE USER \<adminusername\> WITH PASSWORD '\<adminuserpwd\>' WITH ALL PRIVILEGES"`

Once you have started both docker containers and Influx and Grafana are running you need to configure the Jenkins Plugin according to your settings.

## Pipeline configuration

To setup your Jenkins you need to do two configuration steps:

1. Configure Jenkins (via Manage Jenkins)
1. Adapt pipeline configuration

### Configure Jenkins

Once the plugin is available in your Jenkins:

- go to "Manage Jenkins" > "Configure System" > scroll down to section "influxdb target"
- maintain Influx data

!!! note "Jenkins as a Service"
    For Jenkins as a Service instances this is already preset to the local InfluxDB with the name `jenkins`. In this case there is not need to do any additional configuration.

### Adapt pipeline configuration

You need to define the influxDB server in your pipeline as it is defined in the InfluxDb plugin configuration (see above).

```properties
influxDBServer=jenkins
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Example

```groovy
influxWriteData script: this
```

## Work with InfluxDB and Grafana

You can access your **Grafana** via Web-UI: &lt;host of your grafana(-docker)&gt;:&lt;port3000&gt;
(or another port in case you have defined another one when starting your docker)

As a first step you need to add your InfluxDB as Data source to your Grafana:

- Login as user admin (PW as defined when starting your docker)
- in the navigation go to data sources -> add data source:
  - name
  - type: InfluxDB
  - Url: `http://<host of your InfluxDB server>:<port>`
  - Access: direct (not via proxy)
  - database: `<name of the DB as specified above>`
  - User: `<name of the admin user as specified in step above>`
  - Password: `<password of the admin user as specified in step above>`

!!! note "Jenkins as a Service"
    For Jenkins as a Service the data source configuration is already available.

    Therefore no need to go through the data source configuration step unless you want to add additional data sources.

## Data collected in InfluxDB

The Influx plugin collects following data in the project "Piper" context:

- All data as per default [InfluxDB plugin capabilities](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin)
- Additional data collected via `InfluxData.addField(measurement, key, value)`

!!! note "Add custom information to your InfluxDB"
    You can simply add custom data collected during your pipeline runs via available data objects.
    Example:

    ```groovy
    //add data to measurement jenkins_custom_data - value can be a String or a Number
    commonPipelineEnvironment.setInfluxCustomDataProperty('myProperty', 2018)
    ```

### Collected InfluxDB measurements

Measurements are potentially pre-fixed - see parameter `influxPrefix` above.

| Measurement name | data column | description |
| ---------------- | ----------- | ----------- |
| **All measurements** |<ul><li>build_number</li><li>project_name</li></ul>| All below measurements will have these columns. <br />Details see [InfluxDB plugin documentation](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin)|
| jenkins_data | <ul><li>build_result</li><li>build_time</li><li>last_successful_build</li><li>tests_failed</li><li>tests_skipped</li><li>tests_total</li><li>...</li></ul> | Details see [InfluxDB plugin documentation](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin)|
| cobertura_data | <ul><li>cobertura_branch_coverage_rate</li><li>cobertura_class_coverage_rate</li><li>cobertura_line_coverage_rate</li><li>cobertura_package_coverage_rate</li><li>...</li></ul>  | Details see [InfluxDB plugin documentation](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin) |
| jacoco_data | <ul><li>jacoco_branch_coverage_rate</li><li>jacoco_class_coverage_rate</li><li>jacoco_instruction_coverage_rate</li><li>jacoco_line_coverage_rate</li><li>jacoco_method_coverage_rate</li></ul>  | Details see [InfluxDB plugin documentation](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin) |
| performance_data | <ul><li>90Percentile</li><li>average</li><li>max</li><li>median</li><li>min</li><li>error_count</li><li>error_percent</li><li>...</li></ul> | Details see [InfluxDB plugin documentation](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin) |
| sonarqube_data | <ul><li>blocker_issues</li><li>critical_issues</li><li>info_issues</li><li>major_issues</li><li>minor_issues</li><li>lines_of_code</li><li>...</li></ul> | Details see [InfluxDB plugin documentation](https://wiki.jenkins.io/display/JENKINS/InfluxDB+Plugin) |
| jenkins_custom_data | project "Piper" fills following colums by default: <br /><ul><li>build_result</li><li>build_result_key</li><li>build_step (->step in case of error)</li><li>build_error (->error message in case of error)</li></ul> | filled by `commonPipelineEnvironment.setInfluxCustomDataProperty()` |
| pipeline_data | Examples from the project "Piper" templates:<br /><ul><li>build_duration</li><li>opa_duration</li><li>deploy_test_duration</li><li>deploy_test_duration</li><li>fortify_duration</li><li>release_duration</li><li>...</li></ul>| filled by step [`measureDuration`](durationMeasure.md) using parameter `measurementName`|
| step_data | Considered, e.g.:<br /><ul><li>build_url</li><li>bats</li><li>checkmarx</li><li>fortify</li><li>gauge</li><li>nsp</li><li>snyk</li><li>sonar</li><li>...</li></ul>| filled by `InfluxData.addField('step_data', key, value)` |

### Examples for InfluxDB queries which can be used in Grafana

!!! caution "Project Names containing dashes (-)"
    The InfluxDB plugin replaces dashes (-) with underscores (\_).

    Please keep this in mind when specifying your project_name for a InfluxDB query.

#### Example 1: Select last 10 successful builds

```sql
select top(build_number,10), build_result from jenkins_data WHERE build_result = 'SUCCESS'
```

#### Example 2: Select last 10 step names of failed builds

```sql
select top(build_number,10), build_result, build_step from jenkins_custom_data WHERE build_result = 'FAILURE'
```

#### Example 3: Select build duration of step for a specific project

```sql
select build_duration / 1000 from "pipeline_data" WHERE project_name='PiperTestOrg_piper_test_master'
```

#### Example 4: Get transparency about successful/failed steps for a specific project

```sql
select top(build_number,10) AS "Build", build_url, build_quality, fortify, gauge, vulas, opa from step_data WHERE project_name='PiperTestOrg_piper_test_master'
```

!!! note
    With this query you can create transparency about which steps ran successfully / not successfully in your pipeline and which ones were not executed at all.

    By specifying all the steps you consider relevant in your select statement it is very easy to create this transparency.
