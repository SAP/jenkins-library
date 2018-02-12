import org.junit.Rule
import org.junit.Before
import org.junit.Test
import org.junit.rules.ExpectedException

import com.sap.piper.Utils


class UtilsTest {

    @Rule
    public ExpectedException thrown = new ExpectedException().none()

    private utils = new Utils()
    private parameters


    @Before
    void setup() {

        parameters = [:]
    }

    @Test
    void noValueGetMandatoryParameterTest() {

        thrown.expect(Exception)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR test")

        utils.getMandatoryParameter(parameters, 'test', null)
    }

    @Test
    void defaultValueGetMandatoryParameterTest() {

        assert  utils.getMandatoryParameter(parameters, 'test', 'default') == 'default'
    }

    @Test
    void valueGetmandatoryParameterTest() {

        parameters.put('test', 'value')

        assert utils.getMandatoryParameter(parameters, 'test', null) == 'value'
    }
}
