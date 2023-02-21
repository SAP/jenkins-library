package aakaas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

const wildCard string = "NEXT"
const statusFilterCV string = "DeliveryStatus eq 'R'"
const statusFilterPV string = "DeliveryStatus eq 'T' or DeliveryStatus eq 'P'"

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
		return errors.New("No Component/Product Name provided")
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
	v.Version = dottedVersionString
	return nil
}

func (v *versionable) resolveNext(statusFilter string) error {

	switch strings.Count(v.Version, wildCard) {
	case 0:
		return nil
	case 1:
		log.Entry().Info("Wildcard detected in dotted-version-string. Looking up highest existing package in AAKaaS...")
		var err error
		switch wildCard {
		case v.TechRelease:
			err = v.resolveRelease(statusFilter)
		case v.TechSpLevel:
			err = v.resolveSpLevel(statusFilter)
		case v.TechPatchLevel:
			err = v.resolvePatchLevel(statusFilter)
		}
		if err != nil {
			return err
		}
		if v.Version, err = v.getDottedVersionString(); err != nil {
			return err
		}
	default:
		return errors.New("The dotted-version-string must contain only one wildcard " + wildCard)
	}

	return nil
}

func (v *versionable) resolveRelease(statusFilter string) error {
	filter := "Name eq '" + v.Name + "' and TechSpLevel eq '0000' and TechPatchLevel eq '0000' and ( " + statusFilter + " )"
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

func (v *versionable) resolveSpLevel(statusFilter string) error {
	filter := "Name eq '" + v.Name + "' and TechRelease eq '" + v.TechRelease + "' and TechPatchLevel eq '0000'  and ( " + statusFilter + " )"
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

func (v *versionable) resolvePatchLevel(statusFilter string) error {
	filter := "Name eq '" + v.Name + "' and TechRelease eq '" + v.TechRelease + "' and TechSpLevel eq '" + v.TechSpLevel + "' and ( " + statusFilter + " )"
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
	values.Set("$select", "Name,Version,TechRelease,TechSpLevel,TechPatchLevel,Namespace") //Namespace needed otherwise empty result - will be fixed by OCS shortly
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
	log.Entry().Infof("... looked up highest existing package in AAKaaS of the codeline: %s.%s.%s", result.TechRelease, result.TechSpLevel, result.TechPatchLevel)
	return &result, nil
}

func (v *versionable) getDottedVersionString() (string, error) {
	var spLevelAsnumber int
	var patchLevelAsNumber int
	var err error
	if spLevelAsnumber, err = strconv.Atoi(v.TechSpLevel); err != nil {
		return "", err
	}
	if patchLevelAsNumber, err = strconv.Atoi(v.TechPatchLevel); err != nil {
		return "", err
	}
	dottedVersionString := strings.Join([]string{v.TechRelease, strconv.Itoa(spLevelAsnumber), strconv.Itoa(patchLevelAsNumber)}, ".")
	return dottedVersionString, nil
}
