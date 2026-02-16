package whitesource

import (
	"fmt"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

type whitesourcePoller interface {
	GetProjectByToken(projectToken string) (Project, error)
}

// BlockUntilReportsAreReady polls the WhiteSource system for all projects known to the Scan and blocks
// until their LastUpdateDate time stamp is from within the last 20 seconds.
func (s *Scan) BlockUntilReportsAreReady(sys whitesourcePoller) error {
	for _, project := range s.ScannedProjects() {
		if err := pollProjectStatus(project.Token, s.ScanTime(project.Name), sys); err != nil {
			return err
		}
	}
	return nil
}

type pollOptions struct {
	scanTime         time.Time
	maxAge           time.Duration
	timeBetweenPolls time.Duration
	maxWaitTime      time.Duration
}

// pollProjectStatus polls project LastUpdateDate until it reflects the most recent scan
func pollProjectStatus(projectToken string, scanTime time.Time, sys whitesourcePoller) error {
	options := pollOptions{
		scanTime:         scanTime,
		maxAge:           20 * time.Second,
		timeBetweenPolls: 20 * time.Second,
		maxWaitTime:      30 * time.Minute,
	}
	return blockUntilProjectIsUpdated(projectToken, sys, options)
}

// blockUntilProjectIsUpdated polls the project LastUpdateDate until it is newer than the given time stamp
// or no older than maxAge relative to the given time stamp.
func blockUntilProjectIsUpdated(projectToken string, sys whitesourcePoller, options pollOptions) error {
	startTime := time.Now()
	for {
		project, err := sys.GetProjectByToken(projectToken)
		if err != nil {
			return err
		}

		if project.LastUpdateDate == "" {
			log.Entry().Infof("last updated time missing from project metadata, retrying")
		} else {
			lastUpdatedTime, err := time.Parse(DateTimeLayout, project.LastUpdateDate)
			if err != nil {
				return fmt.Errorf("failed to parse last updated time (%s) of Whitesource project: %w",
					project.LastUpdateDate, err)
			}
			age := options.scanTime.Sub(lastUpdatedTime)
			if age < options.maxAge {
				//done polling
				break
			}
			log.Entry().Infof("time since project was last updated %v > %v, polling status...", age, options.maxAge)
		}

		if time.Now().Sub(startTime) > options.maxWaitTime {
			return fmt.Errorf("timeout while waiting for Whitesource scan results to be reflected in service")
		}

		time.Sleep(options.timeBetweenPolls)
	}
	return nil
}
