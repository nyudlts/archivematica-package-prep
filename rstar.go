package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	client  = &http.Client{}
	secrets *Secrets
)

type Secrets struct {
	RSTAR_RO_USER     string
	RSTAR_RO_PASSWORD string
	RSTAR_URI         string
}

func getRStarUUID(partnerCall string) (*string, error) {

	partner, coll := getPartnerAndCall(partnerCall)

	var err error
	secrets, err = readSecretsFile()
	if err != nil {
		return nil, err
	}

	partnerUUID, err := getPartnerUUID(partner)
	if err != nil {
		return nil, err
	}

	collUUID, err := getCollUUID(partnerUUID, coll)
	if err != nil {
		return nil, err
	}

	time.Sleep(pause)
	fmt.Printf("  Validating R* ID: ")
	if uuidMatcher.MatchString(collUUID) != true {
		return nil, fmt.Errorf("%s is not a valid uuid", collUUID)
	}

	return &collUUID, nil
}

func getPartnerUUID(partner string) (string, error) {
	endpoint := fmt.Sprintf("/partners?code=%s", partner)
	bodyJson, err := httpRequest(endpoint)
	if err != nil {
		return "", err
	}
	m := *bodyJson
	return m[0]["id"], nil
}

func getCollUUID(partnerUUID string, coll string) (string, error) {
	endpoint := fmt.Sprintf("/partners/%s/colls?code=%s", partnerUUID, coll)
	bodyJson, err := httpRequest(endpoint)
	if err != nil {
		return "", err
	}
	m := *bodyJson
	return m[0]["id"], nil
}

func httpRequest(endpoint string) (*[]map[string]string, error) {
	url := fmt.Sprintf("%s%s", secrets.RSTAR_URI, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(secrets.RSTAR_RO_USER, secrets.RSTAR_RO_PASSWORD)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyJson := []map[string]string{}
	err = json.Unmarshal(body, &bodyJson)
	if err != nil {
		return nil, err
	}

	return &bodyJson, nil
}

func getPartnerAndCall(partnerCall string) (string, string) {
	parts := strings.Split(partnerCall, "/")
	return parts[0], parts[1]
}

func readSecretsFile() (*Secrets, error) {
	secretsFile, err := os.Open(".secrets")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(secretsFile)
	secrets := Secrets{}
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), ": ")
		if split[0] == "RSTAR_RO_USER" {
			secrets.RSTAR_RO_USER = split[1]
		}

		if split[0] == "RSTAR_RO_PASSWORD" {
			secrets.RSTAR_RO_PASSWORD = split[1]
		}

		if split[0] == "RSTAR_URI" {
			secrets.RSTAR_URI = split[1]
		}
	}

	if secrets.RSTAR_RO_PASSWORD == "" || secrets.RSTAR_RO_USER == "" {
		return nil, fmt.Errorf("SECRETS NOT PARSED CORRECTLY %v", secrets)
	}

	return &secrets, nil
}
