package util

import org.hamcrest.BaseMatcher
import org.hamcrest.Description

class CommandLineMatcher extends BaseMatcher {

    String prolog
    Set<String> args = (Set) []
    Set<MapEntry> opts = (Set) []

    String hint = ''

    CommandLineMatcher hasProlog(prolog) {
        this.prolog = prolog
        return this
    }

    CommandLineMatcher hasDoubleQuotedOption(String key, String value) {
        hasOption(key, "\"${value}\"")
        return this
    }

    CommandLineMatcher hasSingleQuotedOption(String key, String value) {
        hasOption(key, "\'${value}\'")
        return this
    }

    CommandLineMatcher hasOption(String key, String value) {
        this.opts.add(new MapEntry(key, value))
        return this
    }

    CommandLineMatcher hasSnippet(String snippet) {
        this.args.add(snippet)
        return this
    }

    CommandLineMatcher hasArgument(String arg) {
        this.args.add(arg)
        return this
    }

    @Override
    boolean matches(Object o) {

        for (String cmd : o) {

            hint = ''
            boolean matches = true

            if (!cmd.matches(/${prolog}.*/)) {
                hint = "A command line starting with \'${prolog}\'."
                matches = false
            }

            for (MapEntry opt : opts) {
                if (!cmd.matches(/.*[\s]*-${opt.key}[\s]*${opt.value}.*/)) {
                    hint = "A command line containing option \'${opt.key}\' with value \'${opt.value}\'"
                    matches = false
                }
            }

            for (String arg : args) {
                if (!cmd.matches(/.*[\s]*${arg}[\s]*.*/)) {
                    hint = "A command line having argument/snippet '${arg}'."
                    matches = false
                }
            }

            if (matches)
                return true
        }

        return false
    }

    @Override
    public void describeTo(Description description) {
        description.appendText(hint)
    }
}
