#!groovy
package util

import groovy.json.JsonBuilder
import groovy.json.JsonSlurper
import hudson.tasks.junit.TestResult
import org.yaml.snakeyaml.Yaml

/**
 * This is a Helper class for mocking.
 *
 * It can be used to load test data or to mock Jenkins or Maven specific objects. 
 **/

class MockHelper {

    /**
     * load properties from resources for mocking return value of readProperties method
     * @param path to properties
     * @return properties file
     */
    Properties loadProperties( String path ){
        Properties p = new Properties()
        File pFile = new File( path )
        p.load( pFile.newDataInputStream() )
        return p
    }

    /**
     * load JSON from resources for mocking return value of readJSON method
     * @param path to json file
     * @return json file
     */
    Object loadJSON( String path ){
        def js = new JsonSlurper()
        def reader = new BufferedReader(new FileReader( path ))
        def j = js.parse(reader)
        return j
    }

    /**
     * load YAML from resources for mocking return value of readYaml method
     * @param path to yaml file
     * @return yaml file
     */
    Object loadYAML( String path ){
        return new Yaml().load(new FileReader(path))
    }

    /**
     * creates HTTP response for mocking return value of httpRequest method
     * @param text - text to parse into json object
     * @return json Object
     */
    Object createResponse( String text ){
        def response = new JsonBuilder(new JsonSlurper().parseText( text ))
        return response
    }

    /**
     * load File from resources for mocking return value of readFile method
     * @param path to file
     * @return File
     */
    File loadFile( String path ){
        return new File( path )
    }

    /**
     * load POM from resources for mocking return value of readMavenPom method
     * @param path to pom file
     * @return Pom class
     */
    MockPom loadPom(String path ){
        return new MockPom( path )
    }

    /**
     * Inner class to mock maven descriptor
     */
    class MockPom {
        def f
        def pom
        MockPom(String path){
            this.f = new File( path )
            if ( f.exists() ){
                this.pom = new XmlSlurper().parse(f)
            }
            else {
                throw new FileNotFoundException( 'Failed to find file: ' + path )
            }
        }
        String getVersion(){
            return pom.version
        }
        String getGroupId(){
            return pom.groupId
        }
        String getArtifactId(){
            return pom.artifactId
        }
        String getPackaging(){
            return pom.packaging
        }
        String getName(){
            return pom.name
        }
    }

    MockBuild loadMockBuild(){
        return new MockBuild()
    }

    MockBuild loadMockBuild(TestResult result){
        return new MockBuild(result)
    }

    /**
     * Inner class to mock Jenkins' currentBuild return object in scripts
     */
    class MockBuild {
        TestResult testResult
        MockBuild(){}
        MockBuild(TestResult result){
            testResult = result
        }
        MockLibrary getAction(Class c){
            println("MockLibrary -> getAction - arg: " + c.toString() )
            return new MockLibrary()
        }

        class MockLibrary {
            MockLibrary(){}
            // return default
            List getLibraries(){
                println("MockLibrary -> getLibraries")
                return [ [name: 'default-library', version: 'default-master', trusted: true] ]
            }
            TestResult getResult() {
                println("MockLibrary -> getResult")
                return testResult
            }
        }
    }
}
