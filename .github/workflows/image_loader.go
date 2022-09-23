package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

const concurrentAccessNumber int64 = 10

var imgListDefined = []string{
	"devxci/mbtci-java11-node14",
	"golang:1",
	"adoptopenjdk/openjdk11:jdk-11.0.11_9-alpine",
	"selenium/standalone-chrome:3.141.59-20210713",
	"paketobuildpacks/builder:0.3.26-base",
	"node:12-slim",
	"nginx:latest",
	"gradle:6-jdk11-alpine",
	"maven:3-openjdk-8-slim",
	"python:3.9",
	"vault:1.4.3",
	"sonatype/nexus:2.14.18-01",
	"getgauge/gocd-jdk-mvn-node",
	"sonatype/nexus3:3.25.1",
	"paketobuildpacks/builder:buildpackless-full",
	"influxdb:2.0",
	"fsouza/fake-gcs-server:1.30.2",
	"registry:2",
	"nekottyo/kustomize-kubeval:kustomizev4",
	"node:lts-stretch",
}

func main() {
	//path := flag.String("path", "../../integration/", "Path")
	//regExp := flag.String("regexp", "TestCNBIntegrationPreserveFiles\b", "Regular expression to run tests")
	flag.Parse()

	ctx := context.Background()
	//list, err := getImageList(ctx, *path)
	//if err != nil {
	//	log.Panicf("getting image list error: %v", err)
	//}
	//log.Println(list)
	err := loadImages(ctx, imgListDefined...)
	if err != nil {
		log.Fatalf("Image(s) setup error: %v\n", err)
	}
	log.Println("Image(s) are loaded successfully")
}

func loadImages(ctx context.Context, images ...string) error {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Errorf("can't initialize docker client: %w", err)
	}
	sem := semaphore.NewWeighted(concurrentAccessNumber)
	wg, _ := errgroup.WithContext(ctx)
	for _, i := range images {
		if err := sem.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("semaphore: %w", err)
		}
		image := i
		wg.Go(func() error {
			defer sem.Release(1)
			img, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
			if err != nil {
				return fmt.Errorf("can't pull docker image: %w", err)
			}
			_, err = cli.ImageLoad(ctx, img, true)
			if err != nil {
				return fmt.Errorf("can't load docker image: %w", err)
			}
			log.Printf("%s loaded", image)
			return nil
		})
	}
	return wg.Wait()
}

func getImageList(ctx context.Context, path string) (map[string]struct{}, error) {
	funcToImagesStruct := struct {
		m map[string]struct{}
		sync.RWMutex
	}{
		m: make(map[string]struct{}),
	}
	sem := semaphore.NewWeighted(concurrentAccessNumber)
	fs, _ := os.ReadDir(path)
	wg, _ := errgroup.WithContext(ctx)
	for _, f := range fs {
		if !strings.Contains(f.Name(), "_test.go") {
			continue
		}
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("semapfore: %w", err)
		}
		file := f
		wg.Go(func() error {
			data, err := os.Open(path + file.Name())
			if err != nil {
				return fmt.Errorf("file opening error: %w", err)
			}
			defer data.Close()
			scanner := bufio.NewScanner(data)
			// through file scan
			for scanner.Scan() {
				ok, err := regexp.MatchString(`^func\b `, scanner.Text())
				if err != nil {
					return fmt.Errorf("match error: %w", err)
				}
				if !ok {
					continue
				}
				funcName := strings.Split(strings.Split(scanner.Text(), " ")[1], "(")[0]
				// inner function search
				for scanner.Scan() {
					ok, _ := regexp.MatchString(`^}\n`, scanner.Text())
					if ok {
						break
					}
					ok, _ = regexp.MatchString(`\bImage:`, scanner.Text())
					if !ok {
						continue
					}
					funcToImagesStruct.Lock()
					funcToImagesStruct.m[fmt.Sprintf("%s=>%s=>%s\n", file.Name(), funcName, strings.Split(scanner.Text(), "\"")[1])] = struct{}{}
					funcToImagesStruct.Unlock()
				}
			}
			sem.Release(1)
			return nil
		})
	}
	err := wg.Wait()
	if err != nil {
		return nil, fmt.Errorf("goroutine error: %w", err)
	}
	return funcToImagesStruct.m, nil
}
