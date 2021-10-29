package maven

import (
	"encoding/xml"
)

type Server struct {
	ID                   string `xml:"id,omitempty"`
	Username             string `xml:"username,omitempty"`
	Password             string `xml:"password,omitempty"`
	PrivateKey           string `xml:"privateKey,omitempty"`
	Passphrase           string `xml:"passphrase,omitempty"`
	FilePermissions      string `xml:"filePermissions,omitempty"`
	DirectoryPermissions string `xml:"directoryPermissions,omitempty"`
	Configuration        string `xml:"configuration,omitempty"`
}

type ServersType struct {
	ServerType []Server `xml:"server,omitempty"`
}

type ActiveProfilesType struct {
	XMLName       xml.Name `xml:"activeProfiles"`
	Text          string   `xml:",chardata"`
	ActiveProfile []string `xml:"activeProfile"`
}

type Settings struct {
	XMLName xml.Name `xml:"http://maven.apache.org/SETTINGS/1.0.0 settings"`
	Text    string   `xml:",chardata"`
	// Xmlns           xml.Attr `xml:"xmlns,attr"`
	//Xmlns           string `xml:"xmlns,attr"`
	Xsi             string `xml:"xmlns:xsi,attr"`
	SchemaLocation  string `xml:"xsi:schemaLocation,attr"`
	LocalRepository string `xml:"localRepository,omitempty"`
	InteractiveMode string `xml:"interactiveMode,omitempty"`
	Offline         string `xml:"offline,omitempty"`
	PluginGroups    struct {
		Text        string `xml:",chardata"`
		PluginGroup string `xml:"pluginGroup,omitempty"`
	} `xml:"pluginGroups,omitempty"`
	Servers ServersType `xml:"servers,omitempty"`
	Mirrors struct {
		Text   string `xml:",chardata"`
		Mirror []struct {
			Text     string `xml:",chardata"`
			ID       string `xml:"id,omitempty"`
			Name     string `xml:"name,omitempty"`
			URL      string `xml:"url,omitempty"`
			MirrorOf string `xml:"mirrorOf,omitempty"`
		} `xml:"mirror,omitempty"`
	} `xml:"mirrors,omitempty"`
	Proxies struct {
		Text  string `xml:",chardata"`
		Proxy []struct {
			Text          string `xml:",chardata"`
			ID            string `xml:"id,omitempty"`
			Active        string `xml:"active,omitempty"`
			Protocol      string `xml:"protocol,omitempty"`
			Host          string `xml:"host,omitempty"`
			Port          string `xml:"port,omitempty"`
			Username      string `xml:"username,omitempty"`
			Password      string `xml:"password,omitempty"`
			NonProxyHosts string `xml:"nonProxyHosts,omitempty"`
		} `xml:"proxy,omitempty"`
	} `xml:"proxies,omitempty"`
	Profiles struct {
		Text    string `xml:",chardata"`
		Profile []struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,omitempty"`
			// Activation struct {
			// 	Text            string `xml:",chardata"`
			// 	ActiveByDefault string `xml:"activeByDefault,omitempty"`
			// 	Jdk             string `xml:"jdk,omitempty"`
			// 	Os              struct {
			// 		Text    string `xml:",chardata"`
			// 		Name    string `xml:"name,omitempty"`
			// 		Family  string `xml:"family,omitempty"`
			// 		Arch    string `xml:"arch,omitempty"`
			// 		Version string `xml:"version,omitempty"`
			// 	} `xml:"os,omitempty"`
			// 	Property struct {
			// 		Text  string `xml:",chardata"`
			// 		Name  string `xml:"name,omitempty"`
			// 		Value string `xml:"value,omitempty"`
			// 	} `xml:"property,omitempty"`
			// 	File struct {
			// 		Text    string `xml:",chardata"`
			// 		Exists  string `xml:"exists,omitempty"`
			// 		Missing string `xml:"missing,omitempty"`
			// 	} `xml:"file,omitempty"`
			// } `xml:"activation,omitempty"`
			Repositories struct {
				Text       string `xml:",chardata"`
				Repository []struct {
					Text     string `xml:",chardata"`
					ID       string `xml:"id,omitempty"`
					Name     string `xml:"name,omitempty"`
					Releases struct {
						Text           string `xml:",chardata"`
						Enabled        string `xml:"enabled,omitempty"`
						UpdatePolicy   string `xml:"updatePolicy,omitempty"`
						ChecksumPolicy string `xml:"checksumPolicy,omitempty"`
					} `xml:"releases,omitempty"`
					Snapshots struct {
						Text           string `xml:",chardata"`
						Enabled        string `xml:"enabled,omitempty"`
						UpdatePolicy   string `xml:"updatePolicy,omitempty"`
						ChecksumPolicy string `xml:"checksumPolicy,omitempty"`
					} `xml:"snapshots,omitempty"`
					URL    string `xml:"url,omitempty"`
					Layout string `xml:"layout,omitempty"`
				} `xml:"repository,omitempty"`
			} `xml:"repositories,omitempty"`
			PluginRepositories struct {
				Text             string `xml:",chardata"`
				PluginRepository []struct {
					Text     string `xml:",chardata"`
					ID       string `xml:"id,omitempty"`
					Name     string `xml:"name,omitempty"`
					Releases struct {
						Text    string `xml:",chardata"`
						Enabled string `xml:"enabled,omitempty"`
					} `xml:"releases,omitempty"`
					Snapshots struct {
						Text    string `xml:",chardata"`
						Enabled string `xml:"enabled,omitempty"`
					} `xml:"snapshots,omitempty"`
					URL string `xml:"url,omitempty"`
				} `xml:"pluginRepository,omitempty"`
			} `xml:"pluginRepositories,omitempty"`
		} `xml:"profile,omitempty"`
	} `xml:"profiles,omitempty"`
	ActiveProfiles ActiveProfilesType `xml:"activeProfiles,omitempty"`
}
