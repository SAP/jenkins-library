package build

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// Connector : Connector Utility Wrapping http client
type Connector struct {
	Client         piperhttp.Sender
	DownloadClient piperhttp.Downloader
	Header         map[string][]string
	Baseurl        string
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
}

// HTTPSendLoader : combine both interfaces [sender, downloader]
type HTTPSendLoader interface {
	piperhttp.Sender
	piperhttp.Downloader
}

// ******** technical communication calls ********

// GetToken : Get the X-CRSF Token from ABAP Backend for later post
func (conn *Connector) GetToken(appendum string) error {
	url := conn.Baseurl + appendum
	conn.Header["X-CSRF-Token"] = []string{"Fetch"}
	response, err := conn.Client.SendRequest("HEAD", url, nil, conn.Header, nil)
	if err != nil {
		if response == nil {
			return errors.Wrap(err, "Fetching X-CSRF-Token failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errors.Wrapf(err, "Fetching X-CSRF-Token failed: %v", string(errorbody))

	}
	defer response.Body.Close()
	token := response.Header.Get("X-CSRF-Token")
	conn.Header["X-CSRF-Token"] = []string{token}
	return nil
}

// Get : http get request
func (conn Connector) Get(appendum string) ([]byte, error) {
	url := conn.Baseurl + appendum
	response, err := conn.Client.SendRequest("GET", url, nil, conn.Header, nil)
	if err != nil {
		if response == nil || response.Body == nil {
			return nil, errors.Wrap(err, "Get failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Get failed: %v", string(errorbody))

	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

// Post : http post request
func (conn Connector) Post(appendum string, importBody string) ([]byte, error) {
	url := conn.Baseurl + appendum
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
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Post failed: %v", string(errorbody))

	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}

// Download : download a file via http
func (conn Connector) Download(appendum string, downloadPath string) error {
	url := conn.Baseurl + appendum
	err := conn.DownloadClient.DownloadFile(url, downloadPath, nil, nil)
	return err
}

// InitAAKaaS : initialize Connector for communication with AAKaaS backend
func (conn *Connector) InitAAKaaS(aAKaaSEndpoint string, username string, password string, inputclient piperhttp.Sender) {
	conn.Client = inputclient
	conn.Header = make(map[string][]string)
	conn.Header["Accept"] = []string{"application/json"}
	conn.Header["Content-Type"] = []string{"application/json"}
	conn.Header["User-Agent"] = []string{"Piper-abapAddonAssemblyKit/1.0"}

	cookieJar, _ := cookiejar.New(nil)
	conn.Client.SetOptions(piperhttp.ClientOptions{
		Username:  username,
		Password:  password,
		CookieJar: cookieJar,
	})
	conn.Baseurl = aAKaaSEndpoint
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
		Username:  connectionDetails.User,
		Password:  connectionDetails.Password,
		CookieJar: cookieJar,
	})
	conn.Baseurl = connectionDetails.URL

	return nil
}

// UploadSarFile : upload *.sar file
func (conn Connector) UploadSarFile(appendum string, sarFile []byte) error {
	url := conn.Baseurl + appendum
	response, err := conn.Client.SendRequest("PUT", url, bytes.NewBuffer(sarFile), conn.Header, nil)
	if err != nil {
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errors.Wrapf(err, "Upload of SAR file failed: %v", string(errorbody))
	}
	defer response.Body.Close()
	return nil
}

// UploadSarFileInChunks : upload *.sar file in chunks
func (conn Connector) UploadSarFileInChunks(appendum string, fileName string, sarFile []byte) error {
	//Maybe Next Refactoring step to read the file in chunks, too?
	//In case it turns out to be not reliable add a retry mechanism

	url := conn.Baseurl + appendum

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
			if response.Body != nil {
				errorbody, _ := ioutil.ReadAll(response.Body)
				response.Body.Close()
				return errors.Wrapf(err, "Upload of SAR file failed: %v", string(errorbody))
			} else {
				return err
			}
		}

		response.Body.Close()
	}
	return nil
}
