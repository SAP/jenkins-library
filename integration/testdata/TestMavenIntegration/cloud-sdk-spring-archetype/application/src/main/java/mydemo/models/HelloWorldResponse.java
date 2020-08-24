package mydemo.models;

import com.fasterxml.jackson.annotation.JsonProperty;

@SuppressWarnings("PMD.UnusedPrivateField")
public class HelloWorldResponse
{
    @JsonProperty("hello")
    private final String name;

    public HelloWorldResponse( final String name ) {
        this.name = name;
    }
}
