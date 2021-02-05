package mydemo;

import io.restassured.RestAssured;
import org.jboss.arquillian.container.test.api.Deployment;
import org.jboss.arquillian.junit.Arquillian;
import org.jboss.arquillian.test.api.ArquillianResource;
import org.jboss.shrinkwrap.api.spec.WebArchive;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;
import java.net.URL;

import mydemo.HelloWorldServlet;
import com.sap.cloud.sdk.testutil.MockUtil;

import static io.restassured.RestAssured.given;
import static org.assertj.core.api.Assertions.assertThat;

@RunWith( Arquillian.class )
public class HelloWorldServletTest
{
    private static final MockUtil mockUtil = new MockUtil();

    @ArquillianResource
    private URL baseUrl;

    @Deployment
    public static WebArchive createDeployment()
    {
        return TestUtil.createDeployment(HelloWorldServlet.class);
    }

    @BeforeClass
    public static void beforeClass()
    {
        mockUtil.mockDefaults();
    }

    @Before
    public void before()
    {
        RestAssured.baseURI = baseUrl.toExternalForm();
    }

    @Test
    public void testService()
    {
        final String body = given().get("/hello").body().asString();
        assertThat(body).isEqualToIgnoringCase("Hello World!");
    }
}
