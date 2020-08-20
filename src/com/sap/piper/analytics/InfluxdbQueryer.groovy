package com.sap.qod.groovydemo

import com.cloudbees.groovy.cps.NonCPS

import jenkins.model.Jenkins;
import jenkinsci.plugins.influxdb.models.Target;
import jenkinsci.plugins.influxdb.InfluxDbStep.DescriptorImpl

import org.influxdb.InfluxDB;
import org.influxdb.InfluxDBFactory;
import org.influxdb.dto.Query;

import java.util.Date

class InfluxdbQueryer implements Serializable {

    //private static Script script
    private static InfluxDB influxDB

    InfluxdbQueryer(String targetName) {

        //this.script = script
        def target = getInfluxdbTarget(targetName)
        if (target == null) {
            throw new RuntimeException("Target was null!");
        }
        def jenkinsUrl = target.getUrl()
        def jenkinsUsername = target.getUsername()
        def jenkinsDB = target.getDatabase()
        def jenkinsPwd = target.getPassword().getPlainText()
        this.influxDB = InfluxDBFactory.connect(
            jenkinsUrl, jenkinsUsername, target.getPassword().getPlainText()
        )
        this.influxDB.setDatabase(jenkinsDB)

    }

    @NonCPS
    def query(String querystring) {
        /* return a processed single series */
        def rawResult = this.influxDB.query(new Query(querystring))
        return processSeries(rawResult["results"].get(0)['series'].get(0))
    }

    @NonCPS
    def rawQuery(String querystring) {
        // return raw query by influxdb-java
        return this.influxDB.query(new Query(querystring))
    }

    @NonCPS
    def getInfluxdbTarget(String description) {
        // get influxdb target if we configured a jenkins instance with influxdb-plugin
        List<Target> targets = Jenkins.getInstance().getDescriptorByType(DescriptorImpl.class).getTargets();
        for (Target target: targets) {
            String targetInfo = target.getDescription();
            if (targetInfo.equals(description)) {
                return target;
            }
        }
        return null;
    }

    @NonCPS
    def processSeries(seriesMap) {
        def series = [:]
        seriesMap['columns'].each { columnName ->
            series[columnName] = seriesMap['values'].get(0).get(seriesMap['columns'].indexOf(columnName))
        }
        return series
    }

}
