package util

import java.util.List

import groovy.io.FileType

public class StepHelper {

    private static getSteps() {
        List steps = []
        new File('vars').traverse(type: FileType.FILES, maxDepth: 0)
            { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars[\\\/](.*)\.groovy/)[0][1] }
        return steps
    }
}
