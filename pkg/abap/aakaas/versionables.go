package aakaas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/pkg/errors"
)

const wildCard string = "NEXT"

type versionable struct {
	Name           string
	Version        string
	TechRelease    string
	TechSpLevel    string
	TechPatchLevel string

	connector abapbuild.Connector
	queryUrl  string
}

type versionables struct {
	Wrapper struct {
		Vs []versionable `json:"results"`
	} `json:"d"`
}

func (v *versionable) constructVersionable(name string, dottedVersionString string, connector abapbuild.Connector, queryURL string) error {
	if name == "" {
		return errors.New("No Component Name provided")
	}
	subStrings := strings.Split(dottedVersionString, ".")
	if len(subStrings) != 3 {
		return errors.New("Provide a dotted-version-string with 2 '.' [Release.SP.Patch]")
	}
	v.Name = name
	v.TechRelease = subStrings[0]
	v.TechSpLevel = fmt.Sprintf("%04s", subStrings[1])
	v.TechPatchLevel = fmt.Sprintf("%04s", subStrings[2])
	v.connector = connector
	v.queryUrl = queryURL
	return nil
}

func (v *versionable) resolveNext() error {
	switch wildCard {
	case v.TechRelease:
		return v.resolveRelease()
	case v.TechSpLevel:
		return v.resolveSpLevel()
	case v.TechPatchLevel:
		return v.resolvePatchLevel()
	}
	return nil
}

func (v *versionable) resolveRelease() error {
	//take only unrevertable status R/C for packages and T/P for TargetVectors
	filter := "Name eq '" + v.Name + "' and TechSpLevel eq '0000' and TechPatchLevel eq '0000' and ( DeliveryStatus eq 'R' or DeliveryStatus eq 'C' or DeliveryStatus eq 'T' or DeliveryStatus eq 'P' )"
	orderBy := "TechRelease desc"

	if queryResuult, err := v.queryVersion(filter, orderBy); err != nil {
		return err
	} else {
		if newRelease, err := strconv.Atoi(queryResuult.TechRelease); err != nil {
			return err
		} else {
			v.TechRelease = strconv.Itoa(newRelease + 1)
			return nil
		}
	}
}

func (v *versionable) resolveSpLevel() error {
	filter := "Name eq '" + v.Name + "' and TechRelease eq '" + v.TechRelease + "' and TechPatchLevel eq '0000'  and ( DeliveryStatus eq 'R' or DeliveryStatus eq 'C' or DeliveryStatus eq 'T' or DeliveryStatus eq 'P' )"
	orderBy := "TechSpLevel desc"

	if queryResuult, err := v.queryVersion(filter, orderBy); err != nil {
		return err
	} else {
		if newSpLevel, err := strconv.Atoi(queryResuult.TechSpLevel); err != nil {
			return err
		} else {
			v.TechSpLevel = fmt.Sprintf("%04d", newSpLevel+1)
			return nil
		}
	}
}

func (v *versionable) resolvePatchLevel() error {
	filter := "Name eq '" + v.Name + "' and TechRelease eq '" + v.TechRelease + "' and TechSpLevel eq '" + v.TechSpLevel + "' and ( DeliveryStatus eq 'R' or DeliveryStatus eq 'C' or DeliveryStatus eq 'T' or DeliveryStatus eq 'P' )"
	orderBy := "TechPatchLevel desc"

	if queryResuult, err := v.queryVersion(filter, orderBy); err != nil {
		return err
	} else {
		if newPatchLevel, err := strconv.Atoi(queryResuult.TechPatchLevel); err != nil {
			return err
		} else {
			v.TechPatchLevel = fmt.Sprintf("%04d", newPatchLevel+1)
			return nil
		}
	}
}

func (v *versionable) queryVersion(filter string, orderBy string) (*versionable, error) {
	result := versionable{}

	values := url.Values{}
	values.Set("$filter", filter)
	values.Set("$orderby", orderBy)
	values.Set("$select", "Name,Version,TechRelease,TechSpLevel,TechPatchLevel")
	values.Set("$format", "json")
	values.Set("$top", "1")

	requestUrl := v.queryUrl + "?" + values.Encode()

	if body, err := v.connector.Get(requestUrl); err != nil {
		return &result, err
	} else {
		Versions := versionables{}
		if err := json.Unmarshal(body, &Versions); err != nil {
			return &result, errors.Wrap(err, "Unexpected AAKaaS response for Component Version Query: "+string(body))
		}
		switch len(Versions.Wrapper.Vs) {
		case 0:
			result = versionable{
				TechRelease:    "0",
				TechSpLevel:    "0000",
				TechPatchLevel: "0000",
			}
		case 1:
			result = Versions.Wrapper.Vs[0]
		default:
			return &result, errors.New("Unexpected Number of CVs in result: " + fmt.Sprint(len(Versions.Wrapper.Vs)))
		}
	}
	return &result, nil
}
