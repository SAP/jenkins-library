package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/facebookincubator/nvdtools/providers/lib/client"
	"log"
	"net/http"
	"time"
)

type Info struct {
	TotalCount int `json:"total_count"`
	Artifacts  []struct {
		Id                 int       `json:"id"`
		NodeId             string    `json:"node_id"`
		Name               string    `json:"name"`
		SizeInBytes        int       `json:"size_in_bytes"`
		Url                string    `json:"url"`
		ArchiveDownloadUrl string    `json:"archive_download_url"`
		Expired            bool      `json:"expired"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		ExpiresAt          time.Time `json:"expires_at"`
		WorkflowRun        struct {
			Id               int64  `json:"id"`
			RepositoryId     int    `json:"repository_id"`
			HeadRepositoryId int    `json:"head_repository_id"`
			HeadBranch       string `json:"head_branch"`
			HeadSha          string `json:"head_sha"`
		} `json:"workflow_run"`
	} `json:"artifacts"`
}

func main() {
	workflowId := flag.String("workflow", "", "Workflow run id")
	token := flag.String("token", "", "GH token")
	//regExp := flag.String("regexp", "", "Regular expression to run tests")
	flag.Parse()

	request, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.github.com/repos/SAP/jenkins-library/actions/runs/%s/artifacts", *workflowId),
		nil,
	)
	request.Header.Add("Accept", "application/vnd.github+json")
	request.Header.Add("Authorization", fmt.Sprintf("token %s", *token))
	var info Info
	for ; ; time.Sleep(5 * time.Second) {
		resp, _ := client.Default().Do(request)
		if resp.StatusCode != http.StatusOK {
			log.Panicln(resp.Status)
		}
		log.Println(resp.Status)
		json.NewDecoder(resp.Body).Decode(&info)
		if info.TotalCount != 2 {
			log.Println(info)
			log.Println("Pending...")
			continue
		}
		switch info.Artifacts[0].Name {
		case "piper":
		case "integration_tests":
		default:
			log.Panicln("Unknown artifact")
		}
		break
	}
	log.Println("Found")

	//	go func() {
	//		reader := bytes.NewReader(nil)
	//		request, _ := http.NewRequest(
	//			"GET",
	//			fmt.Sprintf("https://api.github.com/repos/SAP/jenkins-library/actions/runs/%s/artifacts", *workflowId),
	//			reader,
	//		)
	//	}()
	//	go func() {
	//		reader := bytes.NewReader(nil)
	//		request, _ := http.NewRequest(
	//			"GET",
	//			fmt.Sprintf("https://api.github.com/repos/SAP/jenkins-library/actions/runs/%s/artifacts", *workflowId),
	//			reader,
	//		)
	//	}()
	//
	//	-H "Accept: application/vnd.github+json" \
	//	-H "Authorization: Bearer <YOUR-TOKEN>" \
	//https://api.github.com/repos/OWNER/REPO/actions/artifacts/ARTIFACT_ID/ARCHIVE_FORMAT

}
