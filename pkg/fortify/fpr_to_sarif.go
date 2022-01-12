package fortify

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"

	"github.com/SAP/jenkins-library/pkg/log"
	FileUtils "github.com/SAP/jenkins-library/pkg/piperutils"
)

type UUID struct {
	XMLName xml.Name `xml:"UUID"`
	Uuid    string   `xml:",internal"`
}

type Build struct {
	XMLName     xml.Name `xml:"Build"`
	Project     string   `xml:"Project"`
	Label       string   `xml:"Label"`
	BuildID     string   `xml:"BuildID"`
	NumberFiles int      `xml:"NumberFiles"`
}

func ConvertFprToSarif(resultFilePath string) error {
	log.Entry().Debug("Extracting FPR.")
	_, err := FileUtils.Unzip(resultFilePath, "result/")
	if err != nil {
		return err
	}
	//File is result/audit.fvdl
	data, err := ioutil.ReadFile("result/audit.fvdl")
	if err != nil {
		return err
	}
	//To read XML data, Unmarshal or Decode can be used. However, Unmarshal is not well-behaved when there are
	//multiple different XML tree roots. This is why a decoder is created from a reader, which allows us to
	//simply run Decode and get all well-formatted XML data for one type.
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	var b Build
	var u UUID

	//Note: it is CRUCIAL that decoding is done in the ORDER OF THE FILE, otherwise decoding WILL FAIL

	decoder.Decode(&u) // Struct u (type UUID) now populated with {{ UUID} 3e8ab4bc-00a7-4772-bafb-918c535b6914}
	decoder.Decode(&b) // Struct b (type Build) now populated with {{ Build} i547985-test https://api.github.com/repos///commits/ 8eefb29f-4441-46ff-b4af-7c17f2b66f9f 35}

	return nil
}
