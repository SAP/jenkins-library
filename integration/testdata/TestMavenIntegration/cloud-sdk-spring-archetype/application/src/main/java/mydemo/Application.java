package mydemo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.boot.web.servlet.ServletComponentScan;
import org.springframework.boot.web.servlet.support.SpringBootServletInitializer;
import org.springframework.context.annotation.ComponentScan;

@SpringBootApplication
@ComponentScan({"com.sap.cloud.sdk", "mydemo"})
@ServletComponentScan({"com.sap.cloud.sdk", "mydemo"})
public class Application extends SpringBootServletInitializer
{
    @Override
    protected SpringApplicationBuilder configure( final SpringApplicationBuilder application )
    {
        return application.sources(Application.class);
    }

    public static void main( final String[] args )
    {
        SpringApplication.run(Application.class, args);
    }
}
