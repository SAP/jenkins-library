package com.sap.piper.analytics

import com.cloudbees.groovy.cps.NonCPS

class InfluxData implements Serializable{

    // each Map in influxCustomDataMap represents a measurement in Influx.
    // Additional measurements can be added as a new Map entry of influxCustomDataMap
    protected Map fields = [jenkins_custom_data: [:], pipeline_data: [:], step_data: [:]]
    // each Map in influxCustomDataMapTags represents tags for certain measurement in Influx.
    // Tags are required in Influx for easier querying data
    protected Map tags = [jenkins_custom_data: [:], pipeline_data: [:], step_data: [:]]

    public Map getFields(){ return fields }
    public Map getTags(){ return tags }

    protected static InfluxData instance

    @NonCPS
    public static InfluxData getInstance(){
        if(!instance) instance = new InfluxData()
        return instance
    }

    public static void addField(String measurement, String key, value) {
        add(getInstance().getFields(), measurement, key, value)
    }

    public static void addTag(String measurement, String key, value) {
        add(getInstance().getTags(), measurement, key, value)
    }

    protected static void add(Map dataMap, String measurement, String field, value) {
        if (!dataMap[measurement]) dataMap[measurement] = [:]
        dataMap[measurement][field] = value
    }

    public static void reset(){
        instance = null
    }

    public static void readFromDisk(script) {

        def influxValues = script.findFiles(glob: '.pipeline/influx/*')

        script.echo "Influx Data: ${getInstance().getFields()} ${getInstance().getTags()}"

        influxValues.each({f ->
            script.echo "Reading ${f}"
            def fileName = f.getName()
            script.echo "File name ${fileName}"
            def parts = fileName.split('.pipeline/influx/')[1].split('/')

            script.echo "File name parts ${parts}"

            if(parts.size() == 3){
                Map dataMap
                def measurement = parts[0]
                def type = parts[1]
                def name = parts[2]

                if(type == 'field'){
                    dataMap = getInstance().getFields()
                }else if(type == 'tag'){
                    dataMap = getInstance().getTags()
                }
                if(dataMap)
                    add(dataMap, measurement, name, script.readFile(f.getPath()))
            }
        })

        script.echo "Influx Data: ${getInstance().getFields()} ${getInstance().getTags()}"
    }
}
