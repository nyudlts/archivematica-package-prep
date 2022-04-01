package main

import (
	"testing"
)

const testColUUID = "b9612d5d-619a-4ceb-b620-d816e4b4340b"
const testPartnerColl = "dlts/test"

func TestRSBEAPI(t *testing.T) {
	t.Run("Test The RSBE API", func(t *testing.T) {
		got, err := getRStarUUID(testPartnerColl)
		if err != nil {
			t.Error(err)
			return
		}
		if got != testColUUID {
			t.Errorf("GOT %s WNATED %s", got, testColUUID)
		}
	})
}
