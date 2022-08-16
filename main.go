package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/open-policy-agent/opa/sdk"
)

var (
	Config = `{
		"services": {
		  "s3": {
			"url": "https://BUCKET.s3.REGION.amazonaws.com",
			"credentials": {
			  "s3_signing": {
				"environment_credentials": {}
			  }
			}
		  }
		},
		"bundles": {
		  "authz": {
			"service": "s3",
			"resource": "bundle.tar.gz"
		  }
		},
        "decision_logs": {
            "console": true
        }
    }`
)

func NewOpaEvaluator(ctx context.Context) (*sdk.OPA, error) {
	// provide the OPA configuration which specifies
	// fetching policy bundles from the mock bundleServer
	// and logging decisions locally to the console
	config := []byte(Config)

	// create an instance of the OPA object
	return sdk.New(ctx, sdk.Options{
		Config: bytes.NewReader(config),
	})
}

type Credentials struct {
	AWS_ACCESS_KEY_ID     string `json:"AWS_ACCESS_KEY_ID"`
	AWS_SECRET_ACCESS_KEY string `json:"AWS_SECRET_ACCESS_KEY"`
	AWS_REGION            string `json:"AWS_REGION"`
}

func main() {
	creds, _ := credentialsFromFile("credentials.json")
	os.Setenv("AWS_ACCESS_KEY_ID", creds.AWS_ACCESS_KEY_ID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", creds.AWS_SECRET_ACCESS_KEY)
	os.Setenv("AWS_REGION", creds.AWS_REGION)

	opa, err := NewOpaEvaluator(context.Background())
	if err != nil {
		fmt.Printf("error in starting opa sdk\n %v", err)
		return
	}

	// JSON input
	input_eks, _ := inputFromFile("input_eks.json")
	input_file, _ := inputFromFile("input_file.json")

	res, err := opa.Decision(context.Background(), sdk.DecisionOptions{
		Path:  "main",
		Input: input_file,
	})
	if err != nil {
		fmt.Printf("error in evaluating input\n %v", err)
		return
	}
	file, _ := json.MarshalIndent(res, "", " ")

	_ = ioutil.WriteFile("output_file.json", file, 0644)

	res, err = opa.Decision(context.Background(), sdk.DecisionOptions{
		Path:  "main",
		Input: input_eks,
	})
	if err != nil {
		fmt.Printf("error in evaluating input\n %v", err)
		return
	}
	file, _ = json.MarshalIndent(res, "", " ")

	_ = ioutil.WriteFile("output_eks.json", file, 0644)
}

func inputFromFile(inputFile string) (interface{}, error) {
	jsonFile, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("error in reading file", err)
	}

	input := map[string]interface{}{}
	json.Unmarshal(jsonFile, &input)
	if err != nil {
		return nil, fmt.Errorf("error in parsing file", err)
	}

	return input, nil
}

func credentialsFromFile(inputFile string) (Credentials, error) {
	jsonFile, err := os.ReadFile(inputFile)
	if err != nil {
		return Credentials{}, fmt.Errorf("error in reading file", err)
	}

	input := Credentials{}
	json.Unmarshal(jsonFile, &input)
	if err != nil {
		return Credentials{}, fmt.Errorf("error in parsing file", err)
	}

	return input, nil
}
