module.exports = function (grunt) {
    'use strict';
    grunt.loadNpmTasks('@sap/grunt-sapui5-bestpractice-build');

    grunt.registerTask('default', [
        'lint',
        'clean',
        'build'
    ]);
};
