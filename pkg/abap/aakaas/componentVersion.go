package aakaas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
)

type ComponentVersion struct {
	Name       string
	Release    string
	SpLevel    string
	PatchLevel string
	connector  abapbuild.Connector
}

type jsonComponentVersions struct {
	Wrapper struct {
		CVs []jsonComponentVersion `json:"results"`
	} `json:"d"`
}

type jsonComponentVersion struct {
	Name           string
	Version        string
	TechRelease    string
	TechSpLevel    string
	TechPatchLevel string
}

const wildCard string = "NEXT"

func NewComponentVersion(name string, dottedVersionString string, connector abapbuild.Connector) (*ComponentVersion, error) {
	if name == "" {
		return nil, errors.New("No Component Name provided")
	}
	subStrings := strings.Split(dottedVersionString, ".")
	if len(subStrings) != 3 {
		return nil, errors.New("Provide a dotted-version-string with 2 '.' [Release|SP|Patch]")
	}
	cv := ComponentVersion{
		Name:       name,
		Release:    subStrings[0],
		SpLevel:    fmt.Sprintf("%04s", subStrings[1]),
		PatchLevel: fmt.Sprintf("%04s", subStrings[2]),
		connector:  connector,
	}

	return &cv, nil
}

func (cv *ComponentVersion) ResolveNext() error {
	switch wildCard {
	case cv.Release:
		if resolvedCV, err := cv.resolveRelease(); err != nil {
			return err
		} else {
			if newRelease, err := strconv.Atoi(resolvedCV.TechRelease); err != nil {
				return err
			} else {
				cv.Release = strconv.Itoa(newRelease + 1)
			}
		}
	case cv.SpLevel:
		if resolvedCV, err := cv.resolveSpLevel(); err != nil {
			return err
		} else {
			if newSpLevel, err := strconv.Atoi(resolvedCV.TechSpLevel); err != nil {
				return err
			} else {
				cv.SpLevel = fmt.Sprintf("%04d", newSpLevel+1)
			}
		}
	case cv.PatchLevel:
		if resolvedCV, err := cv.resolvePatchLevel(); err != nil {
			return err
		} else {
			if newPatchLevel, err := strconv.Atoi(resolvedCV.TechPatchLevel); err != nil {
				return err
			} else {
				cv.PatchLevel = fmt.Sprintf("%04d", newPatchLevel+1)
			}
		}
	}
	return nil
}

func (cv *ComponentVersion) resolveRelease() (*jsonComponentVersion, error) {
	return cv.httpCall("Name eq '"+cv.Name+"' and TechSpLevel eq '0000' and TechPatchLevel eq '0000'", "TechRelease desc")
}

func (cv *ComponentVersion) resolveSpLevel() (*jsonComponentVersion, error) {
	return cv.httpCall("Name eq '"+cv.Name+"' and TechRelease eq '"+cv.Release+"' and TechPatchLevel eq '0000'", "TechSpLevel desc")
}

func (cv *ComponentVersion) resolvePatchLevel() (*jsonComponentVersion, error) {
	return cv.httpCall("Name eq '"+cv.Name+"' and TechRelease eq '"+cv.Release+"' and TechSpLevel eq '"+cv.SpLevel+"'", "TechPatchLevel desc")
}

func (cv *ComponentVersion) httpCall(filter string, orderBy string) (*jsonComponentVersion, error) {
	result := jsonComponentVersion{}
	baseUrl := "/odata/aas_ocs_package/xSSDAxC_Component_Version"

	values := url.Values{}
	values.Set("$filter", filter)
	values.Set("$orderby", orderBy)
	values.Set("$select", "Name,Version,TechRelease,TechSpLevel,TechPatchLevel")
	values.Set("$format", "json")
	values.Set("$top", "1")

	requestUrl := baseUrl + "?" + values.Encode()

	if body, err := cv.connector.Get(requestUrl); err != nil {
		return &result, err
	} else {
		jsonCVs := jsonComponentVersions{}
		if err := json.Unmarshal(body, &jsonCVs); err != nil {
			return &result, errors.Wrap(err, "Unexpected AAKaaS response for Component Version Query: "+string(body))
		}
		switch len(jsonCVs.Wrapper.CVs) {
		case 0:
			result = jsonComponentVersion{
				TechRelease:    "0",
				TechSpLevel:    "0000",
				TechPatchLevel: "0000",
			}
		case 1:
			result = jsonCVs.Wrapper.CVs[0]
		default:
			return &result, errors.New("Unexpected Number of CVs in result: " + fmt.Sprint(len(jsonCVs.Wrapper.CVs)))
		}
	}
	return &result, nil
}
