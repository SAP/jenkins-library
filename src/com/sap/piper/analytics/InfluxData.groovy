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
        def jsonStr = script.sh(script: '''python3 -c "
import os, json, sys
prefix = '.pipeline/influx/'
if not os.path.isdir(prefix):
    print('{}')
    sys.exit(0)
result = {}
for root, dirs, files in os.walk(prefix):
    for fname in sorted(files):
        fpath = os.path.join(root, fname)
        rel = os.path.relpath(fpath, prefix)
        parts = rel.split(os.sep)
        if len(parts) != 3:
            continue
        measurement, typ, name = parts
        if typ not in ('fields', 'tags'):
            continue
        with open(fpath) as f:
            content = f.read().strip()
        if measurement not in result:
            result[measurement] = {}
        if typ not in result[measurement]:
            result[measurement][typ] = {}
        is_json = name.endswith('.json')
        if is_json:
            name = name[:-5]
            try:
                content = json.loads(content)
            except (json.JSONDecodeError, ValueError):
                try:
                    content = json.loads('{\\\"content\\\": ' + content + '}')['content']
                except:
                    pass
        else:
            if content == 'true':
                content = True
            elif content == 'false':
                content = False
        result[measurement][typ][name] = content
print(json.dumps(result))
"''', returnStdout: true).trim()

        if (jsonStr && jsonStr != '{}') {
            def data = script.readJSON(text: jsonStr)
            data.each { String measurement, Map types ->
                types.fields?.each { String name, value ->
                    addField(measurement, name, value)
                }
                types.tags?.each { String name, value ->
                    addTag(measurement, name, value)
                }
            }
        }
    }
}
