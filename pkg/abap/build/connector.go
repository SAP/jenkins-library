package build

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pkcs12"
)

// Connector : Connector Utility Wrapping http client
type Connector struct {
	Client          piperhttp.Sender
	DownloadClient  piperhttp.Downloader
	Header          map[string][]string
	Baseurl         string
	Parameters      url.Values
	MaxRuntime      time.Duration // just as handover parameter for polling functions
	PollingInterval time.Duration // just as handover parameter for polling functions
}

// ConnectorConfiguration : Handover Structure for Connector Creation
type ConnectorConfiguration struct {
	CfAPIEndpoint       string
	CfOrg               string
	CfSpace             string
	CfServiceInstance   string
	CfServiceKeyName    string
	Host                string
	Username            string
	Password            string
	AddonDescriptor     string
	MaxRuntimeInMinutes int
	CertificateNames    []string
	Parameters          url.Values
}

// HTTPSendLoader : combine both interfaces [sender, downloader]
type HTTPSendLoader interface {
	piperhttp.Sender
	piperhttp.Downloader
}

// ******** technical communication calls ********

// GetToken : Get the X-CRSF Token from ABAP Backend for later post
func (conn *Connector) GetToken(appendum string) error {
	url := conn.createUrl(appendum)
	conn.Header["X-CSRF-Token"] = []string{"Fetch"}
	response, err := conn.Client.SendRequest("HEAD", url, nil, conn.Header, nil)
	if err != nil {
		if response == nil {
			return errors.Wrap(err, "Fetching X-CSRF-Token failed")
		}
		defer response.Body.Close()
		errorbody, _ := io.ReadAll(response.Body)
		return errors.Wrapf(err, "Fetching X-CSRF-Token failed: %v", extractErrorStackFromJsonData(errorbody))

	}
	defer response.Body.Close()
	token := response.Header.Get("X-CSRF-Token")
	conn.Header["X-CSRF-Token"] = []string{token}
	log.RegisterSecret(token)

	log.Entry().Debug("response headers:")
	for key, value := range response.Header {
		log.Entry().Debug(key)
		if strings.HasPrefix(key, "SAP_SESSIONID_") {
			log.RegisterSecret(value[0])
			log.Entry().Debug("... registered")
		}
		if key == "Set-Cookie" {
			log.Entry().Debug(">> cookies:")
			for _, cookie := range value {
				log.Entry().Debug(cookie)
			}
			log.Entry().Debug("<< cookies:")
		}
	}

	log.Entry().Debug("conn headers:")
	for key, value := range conn.Header {
		log.Entry().Debug(key)
		if strings.HasPrefix(key, "SAP_SESSIONID_") {
			log.RegisterSecret(value[0])
			log.Entry().Debug("... registered")
		}
	}

	return nil
}

// Get : http get request
func (conn Connector) Get(appendum string) ([]byte, error) {
	url := conn.createUrl(appendum)
	response, err := conn.Client.SendRequest("GET", url, nil, conn.Header, nil)
	if err != nil {
		if response == nil || response.Body == nil {
			return nil, errors.Wrap(err, "Get failed")
		}
		defer response.Body.Close()
		errorbody, _ := io.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Get failed: %v", extractErrorStackFromJsonData(errorbody))

	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return body, err
}

// Post : http post request
func (conn Connector) Post(appendum string, importBody string) ([]byte, error) {
	url := conn.createUrl(appendum)
	var response *http.Response
	var err error
	if importBody == "" {
		response, err = conn.Client.SendRequest("POST", url, nil, conn.Header, nil)
	} else {
		response, err = conn.Client.SendRequest("POST", url, bytes.NewBuffer([]byte(importBody)), conn.Header, nil)
	}
	if err != nil {
		if response == nil {
			return nil, errors.Wrap(err, "Post failed")
		}
		defer response.Body.Close()
		errorbody, _ := io.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Post failed: %v", extractErrorStackFromJsonData(errorbody))

	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return body, err
}

// Download : download a file via http
func (conn Connector) Download(appendum string, downloadPath string) error {
	url := conn.createUrl(appendum)
	err := conn.DownloadClient.DownloadFile(url, downloadPath, nil, nil)
	return err
}

// create url
func (conn Connector) createUrl(appendum string) string {
	myUrl := conn.Baseurl + appendum
	if len(conn.Parameters) == 0 {
		return myUrl
	}
	myUrl = myUrl + "?" + conn.Parameters.Encode()
	return myUrl
}

// InitAAKaaS : initialize Connector for communication with AAKaaS backend
func (conn *Connector) InitAAKaaS(aAKaaSEndpoint string, username string, password string, inputclient piperhttp.Sender, originHash string, certFile string, certPass string) error {
	conn.Client = inputclient
	conn.Header = make(map[string][]string)
	conn.Header["Accept"] = []string{"application/json"}
	conn.Header["Content-Type"] = []string{"application/json"}
	conn.Header["User-Agent"] = []string{"Piper-abapAddonAssemblyKit/1.0"}
	if originHash != "" {
		conn.Header["build-config-token"] = []string{originHash}
		log.Entry().Info("Origin info for restricted scenario added")
	}

	cookieJar, _ := cookiejar.New(nil)
	conn.Baseurl = aAKaaSEndpoint

	tlsCertificates, err := conn.handleLogonCertificate(certFile, certPass)
	if err != nil {
		return errors.Wrap(err, "Handling certificates for client logon failed")
	}

	conn.Client.SetOptions(piperhttp.ClientOptions{
		Username:         username,
		Password:         password,
		CookieJar:        cookieJar,
		Certificates:     tlsCertificates,
		TransportTimeout: 15 * time.Minute, //Usually ABAP Backend has timeout of 10min, let them interrupt the connection...
	})

	if tlsCertificates != nil {
		log.Entry().Info("Logon procedure: via Certificate")
		return nil
	} else if username == "" || password == "" {
		return errors.New("username/password for AAKaaS must not be initial") //leads to redirect to login page which causes HTTP200 instead of HTTP401 and thus side effects
	} else {
		log.Entry().Info("Logon procedure: via Password")
		return nil
	}
}

func (conn *Connector) handleLogonCertificate(certFile, certPass string) ([]tls.Certificate, error) {
	var tlsCertificate tls.Certificate
	if certFile != "" && certPass != "" {
		certFileInBytes, err := base64.StdEncoding.DecodeString(certFile)
		if err != nil {
			return nil, errors.Wrap(err, "Base64 decoding of certificate File string failed")
		}

		pemBlocks, err := pkcs12.ToPEM(certFileInBytes, certPass)
		if err != nil {
			return nil, errors.Wrap(err, "Decoding certificate File from PKCS12 to PEM failed")
		}

		var key []byte
		var userCertificate []byte

		for _, pemBlock := range pemBlocks {
			if pemBlock.Type == "PRIVATE KEY" {
				key = pem.EncodeToMemory(pemBlock)
			}

			if pemBlock.Type == "CERTIFICATE" {
				var tempCert, err = x509.ParseCertificate(pemBlock.Bytes)
				if err != nil {
					return nil, errors.Wrap(err, "Parsing x509 Certificate from PEM Block failed")
				}

				if tempCert.IsCA == false { //We ignore the 2 additional CA Certificates
					userCertificate = pem.EncodeToMemory(pemBlock)
				}
			}
		}

		tlsCertificate, err = tls.X509KeyPair(userCertificate, key)
		if err != nil {
			return nil, errors.Wrap(err, "Creating x509 Key Pair failed")
		}

		return []tls.Certificate{tlsCertificate}, nil

	} else {
		return nil, nil
	}
}

// InitBuildFramework : initialize Connector for communication with ABAP SCP instance
func (conn *Connector) InitBuildFramework(config ConnectorConfiguration, com abaputils.Communication, inputclient HTTPSendLoader) error {
	conn.Client = inputclient
	conn.Header = make(map[string][]string)
	conn.Header["Accept"] = []string{"application/json"}
	conn.Header["Content-Type"] = []string{"application/json"}

	conn.DownloadClient = inputclient
	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{TransportTimeout: 20 * time.Second})
	// Mapping for options
	subOptions := abaputils.AbapEnvironmentOptions{}
	subOptions.CfAPIEndpoint = config.CfAPIEndpoint
	subOptions.CfServiceInstance = config.CfServiceInstance
	subOptions.CfServiceKeyName = config.CfServiceKeyName
	subOptions.CfOrg = config.CfOrg
	subOptions.CfSpace = config.CfSpace
	subOptions.Host = config.Host
	subOptions.Password = config.Password
	subOptions.Username = config.Username

	// Determine the host, user and password, either via the input parameters or via a cloud foundry service key
	connectionDetails, err := com.GetAbapCommunicationArrangementInfo(subOptions, "/sap/opu/odata/BUILD/CORE_SRV")
	if err != nil {
		return errors.Wrap(err, "Parameters for the ABAP Connection not available")
	}

	conn.DownloadClient.SetOptions(piperhttp.ClientOptions{
		Username: connectionDetails.User,
		Password: connectionDetails.Password,
	})
	cookieJar, _ := cookiejar.New(nil)
	conn.Client.SetOptions(piperhttp.ClientOptions{
		Username:     connectionDetails.User,
		Password:     connectionDetails.Password,
		CookieJar:    cookieJar,
		TrustedCerts: config.CertificateNames,
	})
	conn.Baseurl = connectionDetails.URL
	conn.Parameters = config.Parameters

	return nil
}

// UploadSarFile : upload *.sar file
func (conn Connector) UploadSarFile(appendum string, sarFile []byte) error {
	url := conn.createUrl(appendum)
	response, err := conn.Client.SendRequest("PUT", url, bytes.NewBuffer(sarFile), conn.Header, nil)
	if err != nil {
		defer response.Body.Close()
		errorbody, _ := io.ReadAll(response.Body)
		return errors.Wrapf(err, "Upload of SAR file failed: %v", extractErrorStackFromJsonData(errorbody))
	}
	defer response.Body.Close()
	return nil
}

// UploadSarFileInChunks : upload *.sar file in chunks
func (conn Connector) UploadSarFileInChunks(appendum string, fileName string, sarFile []byte) error {
	//Maybe Next Refactoring step to read the file in chunks, too?
	//In case it turns out to be not reliable add a retry mechanism

	url := conn.createUrl(appendum)

	header := make(map[string][]string)
	header["Content-Disposition"] = []string{"form-data; name=\"file\"; filename=\"" + fileName + "\""}

	//chunkSize := 10000 // 10KB for testing
	//chunkSize := 1000000 //1MB for Testing,
	chunkSize := 10000000 //10MB
	log.Entry().Infof("Upload in chunks of %d bytes", chunkSize)

	sarFileBuffer := bytes.NewBuffer(sarFile)
	fileSize := sarFileBuffer.Len()

	for sarFileBuffer.Len() > 0 {
		startOffset := fileSize - sarFileBuffer.Len()
		nextChunk := bytes.NewBuffer(sarFileBuffer.Next(chunkSize))
		endOffset := fileSize - sarFileBuffer.Len()
		header["Content-Range"] = []string{"bytes " + strconv.Itoa(startOffset) + " - " + strconv.Itoa(endOffset) + " / " + strconv.Itoa(fileSize)}
		log.Entry().Info(header["Content-Range"])

		response, err := conn.Client.SendRequest("POST", url, nextChunk, header, nil)
		if err != nil {
			if response != nil && response.Body != nil {
				errorbody, _ := io.ReadAll(response.Body)
				response.Body.Close()
				return errors.Wrapf(err, "Upload of SAR file failed: %v", extractErrorStackFromJsonData(errorbody))
			} else {
				return err
			}
		}

		response.Body.Close()
	}
	return nil
}
