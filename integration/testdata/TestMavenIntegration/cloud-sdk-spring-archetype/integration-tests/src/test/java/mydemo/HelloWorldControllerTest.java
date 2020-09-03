package mydemo;

import java.io.InputStream;
import org.apache.commons.codec.Charsets;
import org.apache.commons.io.IOUtils;
import org.junit.BeforeClass;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;

import com.sap.cloud.sdk.cloudplatform.thread.ThreadContextExecutor;
import com.sap.cloud.sdk.testutil.MockUtil;

import static java.lang.Thread.currentThread;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@RunWith( SpringRunner.class )
@WebMvcTest
public class HelloWorldControllerTest
{
    private static final MockUtil mockUtil = new MockUtil();

    @Autowired
    private MockMvc mvc;

    @BeforeClass
    public static void beforeClass()
    {
        mockUtil.mockDefaults();
    }

    @Test
    public void test() throws Exception
    {
        final InputStream inputStream = currentThread().getContextClassLoader().getResourceAsStream("expected.json");

        new ThreadContextExecutor().execute(() -> {
            mvc.perform(MockMvcRequestBuilders.get("/hello"))
                    .andExpect(status().isOk())
                    .andExpect(content().json(
                            IOUtils.toString(inputStream, Charsets.UTF_8)));
        });
    }
}
