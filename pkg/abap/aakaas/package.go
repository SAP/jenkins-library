package aakaas

import (
	"encoding/json"
	"net/url"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// PackageStatus : Status of an ABAP delivery package
type PackageStatus string

const (
	// PackageStatusPlanned : Package is Planned
	PackageStatusPlanned PackageStatus = "P"
	// PackageStatusLocked : Package is Locked
	PackageStatusLocked PackageStatus = "L"
	// PackageStatusReleased : Package is Released
	PackageStatusReleased PackageStatus = "R"
	// PackageStatusCreationTriggered : Package was Released but Release procedure is not yet finished
	PackageStatusCreationTriggered PackageStatus = "C"
)

type jsonPackageDeterminePackageForScv struct {
	DeterminePackage struct {
		Package *Package `json:"DeterminePackageForScv"`
	} `json:"d"`
}

type jsonPackage struct {
	Package *Package `json:"d"`
}

// Package : ABAP delivery package
type Package struct {
	abapbuild.Connector
	ComponentName       string
	VersionYAML         string
	PackageName         string        `json:"Name"`
	Type                string        `json:"Type"`
	PredecessorCommitID string        `json:"PredecessorCommitId"`
	CommitID            string        `json:"CommitId"`
	Status              PackageStatus `json:"Status"`
	Namespace           string        `json:"Namespace"`
}

// PackageWithRepository : pack'n repo
type PackageWithRepository struct {
	Package Package
	Repo    abaputils.Repository
}

// InitPackage : initialize package attributes from the repository
func (p *Package) InitPackage(repo abaputils.Repository, conn abapbuild.Connector) {
	p.Connector = conn
	p.ComponentName = repo.Name
	p.VersionYAML = repo.VersionYAML
	p.PackageName = repo.PackageName
	p.Status = PackageStatus(repo.Status)
}

// CopyFieldsToRepo : copy package attributes to the repository
func (p *Package) CopyFieldsToRepo(initialRepo *abaputils.Repository) {
	initialRepo.PackageName = p.PackageName
	initialRepo.PackageType = p.Type
	initialRepo.PredecessorCommitID = p.PredecessorCommitID
	initialRepo.Status = string(p.Status)
	initialRepo.Namespace = p.Namespace
}

// ReserveNext : reserve next delivery package for this software component version
func (p *Package) ReserveNext() error {
	if p.ComponentName == "" || p.VersionYAML == "" {
		return errors.New("Parameters missing. Please provide the name and version of the component")
	}
	log.Entry().Infof("... determining package name and attributes for software component %s version %s", p.ComponentName, p.VersionYAML)
	p.Connector.GetToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/DeterminePackageForScv?Name='" + url.QueryEscape(p.ComponentName) + "'&Version='" + url.QueryEscape(p.VersionYAML) + "'"
	body, err := p.Connector.Post(appendum, "")
	if err != nil {
		return err
	}
	var jPck jsonPackageDeterminePackageForScv
	if err := json.Unmarshal(body, &jPck); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for reserve package: "+string(body))
	}
	p.PackageName = jPck.DeterminePackage.Package.PackageName
	p.Type = jPck.DeterminePackage.Package.Type
	p.PredecessorCommitID = jPck.DeterminePackage.Package.PredecessorCommitID
	p.Status = jPck.DeterminePackage.Package.Status
	p.setNamespace(jPck.DeterminePackage.Package.Namespace)
	p.CommitID = jPck.DeterminePackage.Package.CommitID
	if p.Status == PackageStatusReleased {
		log.Entry().Infof(" => Reservation of package %s not needed as status is already 'released'", p.PackageName)
	} else {
		log.Entry().Infof(" => Reservation of package %s started", p.PackageName)
	}
	return nil
}

// GetPackageAndNamespace : retrieve attributes of the package from AAKaaS
func (p *Package) GetPackageAndNamespace() error {
	appendum := "/odata/aas_ocs_package/OcsPackageSet('" + url.QueryEscape(p.PackageName) + "')"
	body, err := p.Connector.Get(appendum)
	if err != nil {
		return err
	}

	var jPck jsonPackage
	if err := json.Unmarshal(body, &jPck); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for check of package status: "+string(body))
	}

	p.Status = jPck.Package.Status
	p.setNamespace(jPck.Package.Namespace)

	return nil
}

// ChangeStatus : change status of the package in the repository
func (p *Package) ChangeStatus(initialRepo *abaputils.Repository) {
	initialRepo.Status = string(p.Status)
}

// Register : register package in AAKaaS
func (p *Package) Register() error {
	if p.PackageName == "" {
		return errors.New("Parameter missing. Please provide the name of the package which should be registered")
	}
	log.Entry().Infof("Register package %s", p.PackageName)
	p.Connector.GetToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/RegisterPackage?Name='" + url.QueryEscape(p.PackageName) + "'"
	body, err := p.Connector.Post(appendum, "")
	if err != nil {
		return err
	}

	var jPck jsonPackage
	if err := json.Unmarshal(body, &jPck); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for register package: "+string(body))
	}
	p.Status = jPck.Package.Status
	log.Entry().Infof("Package status %s", p.Status)
	return nil
}

// Release : release package in AAKaaS
func (p *Package) Release() error {
	var body []byte
	var err error
	log.Entry().Infof("Release package %s", p.PackageName)
	p.Connector.GetToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/ReleasePackage?Name='" + url.QueryEscape(p.PackageName) + "'"
	body, err = p.Connector.Post(appendum, "")
	if err != nil {
		return err
	}
	var jPck jsonPackage
	if err := json.Unmarshal(body, &jPck); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for release package: "+string(body))
	}
	p.Status = jPck.Package.Status
	return nil
}

// setNamespace
func (p *Package) setNamespace(namespace string) {
	if namespace == "//" {
		p.Namespace = ""
	} else {
		p.Namespace = namespace
	}
}
