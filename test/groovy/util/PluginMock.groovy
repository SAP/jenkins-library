package util

class PluginMock {

    String pluginName

    PluginMock(String pluginName) {
        this.pluginName = pluginName
    }

    String getShortName() {
        return pluginName
    }

    boolean isActive() {
        return !pluginName.isEmpty()
    }
}
