package abaputils

import (
	"reflect"

	"github.com/SAP/jenkins-library/pkg/log"
)

// ObjectSet in form of packages and software components to be checked
type ObjectSet struct {
	PackageNames       []Package            `json:"packages,omitempty"`
	SoftwareComponents []SoftwareComponents `json:"softwarecomponents,omitempty"`
	PackageTrees       []PackageTree        `json:"packagetrees,omitempty"`
	Type               string               `json:"type,omitempty"`
	MultiPropertySet   MultiPropertySet     `json:"multipropertyset"`
	Set                []Set                `json:"set,omitempty"`
}

// MultiPropertySet that can possibly contain any subsets/object of the OSL
type MultiPropertySet struct {
	Type                  string                 `json:"type,omitempty"`
	PackageNames          []Package              `json:"packages,omitempty"`
	PackageTrees          []PackageTree          `json:"packagetrees,omitempty"`
	ObjectTypeGroups      []ObjectTypeGroup      `json:"objecttypegroups,omitempty"`
	ObjectTypes           []ObjectType           `json:"objecttypes,omitempty"`
	Owners                []Owner                `json:"owners,omitempty"`
	ReleaseStates         []ReleaseState         `json:"releasestates,omitempty"`
	Versions              []Version              `json:"versions,omitempty"`
	ApplicationComponents []ApplicationComponent `json:"applicationcomponents,omitempty"`
	SoftwareComponents    []SoftwareComponents   `json:"softwarecomponents,omitempty"`
	TransportLayers       []TransportLayer       `json:"transportlayers,omitempty"`
	Languages             []Language             `json:"languages,omitempty"`
	SourceSystems         []SourceSystem         `json:"sourcesystems,omitempty"`
}

// Set
type Set struct {
	Type          string          `json:"type,omitempty"`
	Set           []Set           `json:"set,omitempty"`
	PackageSet    []PackageSet    `json:"package,omitempty"`
	FlatObjectSet []FlatObjectSet `json:"object,omitempty"`
	ComponentSet  []ComponentSet  `json:"component,omitempty"`
	TransportSet  []TransportSet  `json:"transport,omitempty"`
	ObjectTypeSet []ObjectTypeSet `json:"objecttype,omitempty"`
}

// PackageSet in form of packages to be checked
type PackageSet struct {
	Name               string `json:"name,omitempty"`
	IncludeSubpackages *bool  `json:"includesubpackages,omitempty"`
}

// FlatObjectSet
type FlatObjectSet struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// ComponentSet in form of software components to be checked
type ComponentSet struct {
	Name string `json:"name,omitempty"`
}

// TransportSet in form of transports to be checked
type TransportSet struct {
	Number string `json:"number,omitempty"`
}

// ObjectTypeSet
type ObjectTypeSet struct {
	Name string `json:"name,omitempty"`
}

// Package for MPS
type Package struct {
	Name string `json:"name,omitempty"`
}

// Packagetree for MPS
type PackageTree struct {
	Name string `json:"name,omitempty"`
}

// ObjectTypeGroup
type ObjectTypeGroup struct {
	Name string `json:"name,omitempty"`
}

// ObjectType
type ObjectType struct {
	Name string `json:"name,omitempty"`
}

// Owner
type Owner struct {
	Name string `json:"name,omitempty"`
}

// ReleaseState
type ReleaseState struct {
	Value string `json:"value,omitempty"`
}

// Version
type Version struct {
	Value string `json:"value,omitempty"`
}

// ApplicationComponent
type ApplicationComponent struct {
	Name string `json:"name,omitempty"`
}

// SoftwareComponents
type SoftwareComponents struct {
	Name string `json:"name,omitempty"`
}

// TransportLayer
type TransportLayer struct {
	Name string `json:"name,omitempty"`
}

// Language
type Language struct {
	Value string `json:"value,omitempty"`
}

// SourceSystem
type SourceSystem struct {
	Name string `json:"name,omitempty"`
}

func BuildOSLString(OSLConfig ObjectSet) (objectSetString string) {

	//Build ObjectSets
	s := OSLConfig
	if s.Type == "" {
		s.Type = "multiPropertySet"
	}
	switch s.Type {
	case "multiPropertySet":
		objectSetString += `<osl:objectSet xsi:type="` + s.Type + `" xmlns:osl="http://www.sap.com/api/osl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">`

		if !(reflect.DeepEqual(s.PackageNames, Package{})) || !(reflect.DeepEqual(s.SoftwareComponents, SoftwareComponents{})) {
			//To ensure Scomps and packages can be assigned on this level
			mps := MultiPropertySet{
				PackageNames:       s.PackageNames,
				SoftwareComponents: s.SoftwareComponents,
			}
			objectSetString += writeObjectSetProperties(mps)
		}

		objectSetString += writeObjectSetProperties(s.MultiPropertySet)

		if !(reflect.DeepEqual(s.MultiPropertySet, MultiPropertySet{})) {
			log.Entry().Info("Wrong configuration has been detected: MultiPropertySet has been used. Please note that there is no official documentation for this usage. Please check the step documentation for more information")
		}

		for _, t := range s.Set {
			log.Entry().Infof("Wrong configuration has been detected: %s has been used. This is currently not supported and this set will not be included in this run. Please check the step documentation for more information", t.Type)
		}
		objectSetString += `</osl:objectSet>`

	default:
		log.Entry().Infof("Wrong configuration has been detected: %s has been used. This is currently not supported and this set will not be included in this run. Please check the step documentation for more information", s.Type)
	}

	return objectSetString
}

func writeObjectSetProperties(set MultiPropertySet) (objectSetString string) {
	for _, packages := range set.PackageNames {
		objectSetString += `<osl:package name="` + packages.Name + `"/>`
	}
	for _, packagetrees := range set.PackageTrees {
		objectSetString += `<osl:package name="` + packagetrees.Name + `" includeSubpackages="true"/>`
	}
	for _, objectTypeGroup := range set.ObjectTypeGroups {
		objectSetString += `<osl:objectTypeGroup name="` + objectTypeGroup.Name + `"/>`
	}
	for _, objectType := range set.ObjectTypes {
		objectSetString += `<osl:objectType name="` + objectType.Name + `"/>`
	}
	for _, owner := range set.Owners {
		objectSetString += `<osl:owner name="` + owner.Name + `"/>`
	}
	for _, releaseState := range set.ReleaseStates {
		objectSetString += `<osl:releaseState value="` + releaseState.Value + `"/>`
	}
	for _, version := range set.Versions {
		objectSetString += `<osl:version value="` + version.Value + `"/>`
	}
	for _, applicationComponent := range set.ApplicationComponents {
		objectSetString += `<osl:applicationComponent name="` + applicationComponent.Name + `"/>`
	}
	for _, component := range set.SoftwareComponents {
		objectSetString += `<osl:softwareComponent name="` + component.Name + `"/>`
	}
	for _, transportLayer := range set.TransportLayers {
		objectSetString += `<osl:transportLayer name="` + transportLayer.Name + `"/>`
	}
	for _, language := range set.Languages {
		objectSetString += `<osl:language value="` + language.Value + `"/>`
	}
	for _, sourceSystem := range set.SourceSystems {
		objectSetString += `<osl:sourceSystem name="` + sourceSystem.Name + `"/>`
	}
	return objectSetString
}
