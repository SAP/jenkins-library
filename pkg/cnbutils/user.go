package cnbutils

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

func CnbUserInfo() (int, int, error) {
	uidStr, ok := os.LookupEnv("CNB_USER_ID")
	if !ok {
		return 0, 0, errors.New("environment variable CNB_USER_ID not found")
	}

	gidStr, ok := os.LookupEnv("CNB_GROUP_ID")
	if !ok {
		return 0, 0, errors.New("environment variable CNB_GROUP_ID not found")
	}

	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return 0, 0, err
	}

	gid, err := strconv.Atoi(gidStr)
	if err != nil {
		return 0, 0, err
	}

	return uid, gid, nil
}
