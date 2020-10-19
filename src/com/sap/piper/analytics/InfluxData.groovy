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

    public static void addField(String measurement, String key, def value) {
        add(getInstance().getFields(), measurement, key, value)
    }

    public static void addTag(String measurement, String key, def value) {
        add(getInstance().getTags(), measurement, key, value)
    }

    protected static void add(Map dataMap, String measurement, String field, def value) {
        if (!dataMap[measurement]) dataMap[measurement] = [:]
        dataMap[measurement][field] = value
    }

    public static void reset(){
        instance = null
    }

    public static void readFromDisk(script) {
        script.echo "Transfer Influx data"
        def pathPrefix = '.pipeline/influx/'
        List influxDataFiles = script.findFiles(glob: "${pathPrefix}**")?.toList()

        influxDataFiles.each({f ->
            script.echo "Reading file form disk: ${f}"
            List parts = f.toString().replace(pathPrefix, '')?.split('/')?.toList()

            if(parts?.size() == 3){
                def type = parts?.get(1)

                if(type in ['fields', 'tags']){
                    def fileContent = script.readFile(f.getPath())
                    def measurement = parts?.get(0)
                    def name = parts?.get(2)
                    def value
                    if (name.endsWith(".json")){
                        script.echo "reading JSON content: " + fileContent
                        name = name.replace(".json","")
                        // net.sf.json.JSONSerializer does only handle lists and maps
                        // http://json-lib.sourceforge.net/apidocs/net/sf/json/package-summary.html
                        try{
                            value = script.readJSON(text: fileContent)
                        }catch(net.sf.json.JSONException e){
                            // try to wrap the value in an object and read again
                            if (e.getMessage() == "Invalid JSON String"){
                                value = script.readJSON(text: "{\"content\": ${fileContent}}").content
                            }else{
                                throw e
                            }
                        }
                    }else{
                        // handle boolean values
                        if(fileContent == 'true'){
                            value = true
                        }else if(fileContent == 'false'){
                            value = false
                        }else{
                            value = fileContent
                        }
                    }

                    if(type == 'fields'){
                        addField(measurement, name, value)
                    }else{
                        addTag(measurement, name, value)
                    }
                }
            } else {
                script.echo "skipped, illegal path"
            }
        })
    }
}
