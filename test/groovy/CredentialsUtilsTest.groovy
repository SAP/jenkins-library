import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

import com.sap.piper.CredentialsUtils

import hudson.AbortException

public class CredentialsUtilsTest {

    static messages = []
    static parameters = [:]
    static userRemoteConfigs = []

    static class MyRemoteConfig {
        String url
        String credentialsId
    }

    @Rule
    public ExpectedException thrown = ExpectedException.none()

    private CredentialsUtils examinee

    @Before
    public void setup() {
        parameters.clear()
        messages.clear()
        userRemoteConfigs.clear()

        //
        // Needed since we have tests manipulating the CredentialsUtils meta class
        // on the level of an individual test. Needs to be reset before the next test.
        GroovySystem.metaClassRegistry.removeMetaClass(CredentialsUtils.class)

        CredentialsUtils.metaClass.params = parameters
        CredentialsUtils.metaClass.echo = { s -> messages.add(s)}
        CredentialsUtils.metaClass.scm = [userRemoteConfigs : userRemoteConfigs]

        examinee = new CredentialsUtils()
    }

    @Test
    public void getCredentialsIdFromDefaultJobParameterTest() {
        parameters.put('CREDENTIALS_ID', 'abc')
        assert examinee.getCredentialsIdFromJobParameters() == 'abc'
    }

    @Test
    public void credentialsIdFromJobParameterHasPriorityOverParameterFromSCMTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'xyz'])])
        assert examinee.getCredentialsId('https://example.org/abc.git') == 'xyz'
        parameters.put('CREDENTIALS_ID', 'abc')
        assert examinee.getCredentialsId('https://example.org/abc.git') == 'abc'
        assert messages.find({ m -> m =~ /\[WARNING\] CredentialsId in SCM configuration (.*) differs from credentialsId retrieved from job parameters (.*). CredentialsId from job parameters will be used./})

    }

    @Test
    public void noneValueFromJobParametersIsIgnoredTest() {
        parameters.put('CREDENTIALS_ID', '- none -')
        assert examinee.getCredentialsIdFromJobParameters(~'.*', ~'^- none -$') == null
    }

    @Test
    public void getFilteredValueFromJobParametersIsIgnoredTest() {
        parameters.put('CREDENTIALS_ID', 'abc/123')
        assert examinee.getCredentialsIdFromJobParameters() == 'abc/123'
        assert examinee.getCredentialsIdFromJobParameters(~'^.*(?=\\/)') == 'abc'
    }

    @Test
    public void getCredentialsIdFromNonStandardJobParameterTest() {
        parameters.put('CREDENTIALS_ID', 'must_not_be_returned')
        parameters.put('THE_CREDENTIALS_ID', 'abc')
        assert examinee.getCredentialsIdFromJobParameters( ~'.*', null, 'THE_CREDENTIALS_ID') == 'abc'
    }

    @Test
    public void repositoryNotMaintainedTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: null])])
        assert examinee.getCredentialsIdFromJobSCMConfig('https://example.org/123.git') == null
    }

    @Test
    public void repoFoundButNoCredentialsMaintainedTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: null])])
        assert examinee.getCredentialsIdFromJobSCMConfig('https://example.org/abc.git') == null
    }

    @Test
    public void repoFoundCredentialsMaintainedTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'credId'])])
        assert examinee.getCredentialsIdFromJobSCMConfig('https://example.org/abc.git') == 'credId'
    }

    @Test
    public void credentalsFromJobConfigRepoUrlIsNullTest() {
        thrown.expect(IllegalArgumentException.class)
        thrown.expectMessage('repoUrl was null or empty.')

        examinee.getCredentialsIdFromJobSCMConfig(null)
    }

    @Test
    public void twoReposMaintainedRepoFoundCredentialsMaintainedTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'credIdabc']),
                                                        new MyRemoteConfig([url: 'https://example.org/123.git', credentialsId: 'credId123'])])
        assert examinee.getCredentialsIdFromJobSCMConfig('https://example.org/abc.git') == 'credIdabc'
    }

    @Test
    public void twoReposWithSameURLMaintainedCredentialsFromFirstRepoReturnedTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'credIdabc']),
                             new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'credId123'])])
        assert examinee.getCredentialsIdFromJobSCMConfig('https://example.org/abc.git') == 'credIdabc'
    }

    @Test
    public void resolveCredentialsIdWithGStringTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'credId'])])
        def repoAsString = 'https://example.org/abc.git'
        def repoAsGroovyString = "${repoAsString}"
        assert repoAsGroovyString instanceof GString
        assert examinee.getCredentialsIdFromJobSCMConfig(repoAsGroovyString) == 'credId'
    }

    @Test
    public void resolveCredentialsIdWithJavaLangStringTest() {
        userRemoteConfigs.addAll([new MyRemoteConfig([url: 'https://example.org/abc.git', credentialsId: 'credId'])])
        def repo = 'https://example.org/abc.git'
        assert repo instanceof String
        assert examinee.getCredentialsIdFromJobSCMConfig(repo) == 'credId'
    }

    @Test
    public void getCredentialsScmNotPresentTestJobParameterPresent() {

        missingScm()
        parameters.put('CREDENTIALS_ID', 'abc')

        def credentialsId = examinee.getCredentialsId('https://example.org/abc.git')

        assert credentialsId == 'abc'
        assert messages.find( {m -> m == '[INFO] hudson.AbortException caught while retrieving credentialsId from SCM configuration. This does not indicate any problem in case the pipeline script is inlined in the job configuration.'})
    }

    @Test
    public void getCredentialsScmNotPresentTestJobParameterNotPresent() {

        missingScm()

        def credentialsId = examinee.getCredentialsId('https://example.org/abc.git')

        assert credentialsId == null
        assert messages.find( {m -> m == '[INFO] hudson.AbortException caught while retrieving credentialsId from SCM configuration. This does not indicate any problem in case the pipeline script is inlined in the job configuration.'})
    }

    @Test
    public void getCredentialsFromJobScmScmNotPresentTest() {

        thrown.expect(AbortException.class)
        thrown.expectMessage('SCM not found.')

        missingScm()

        examinee.getCredentialsIdFromJobSCMConfig('https://example.org/abc.git')
    }

    private static void missingScm() {
        CredentialsUtils.metaClass.retrieveScm = { -> throw new AbortException('SCM not found.') }
    }
 }
