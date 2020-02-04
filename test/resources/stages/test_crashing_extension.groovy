void call(Map params) {
    params.originalStage()
    String somePath = 'path/to/file'
    int index = somePath.indexOf('fi1e') // Index is not what you think it is...
    String fileName = somePath.substring(index) // ...Crash!
    echo "File name is ${fileName}"
}
return this
