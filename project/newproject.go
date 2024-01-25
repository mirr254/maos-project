package project

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Project struct {
	ProjectName  string `json:"project_name"`
	AwsRegion    string `json:"aws_region"`

}

type ProjectResponse struct {
	ProjectName string `json:"project_name"`
}

var projectRootDir string = "../"

func PulumiUp(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-type", "application/json")

	var project Project
	err := json.NewDecoder(req.Body).Decode(&project)
	if err != nil {
		w.WriteHeader(304)
		fmt.Fprint(w, "Failed to parse project args")
	}

	_, err = CreateProject( project.ProjectName, &project.AwsRegion)
	if err != nil {
		w.WriteHeader(304)
		fmt.Fprint(w, err)
	}

	w.WriteHeader(200)

	
}
/*
CreateProject: function is responsible for creating a new project on pulumi dashboard.
A Client can have multiple projects.
params: projectName: This can be client's new project

*/
func CreateProject( projectName string, region *string ) ([]byte, error) {

	var awsregion string
	if region != nil {
        awsregion = *region
	}
	
	pulimiFile, err := os.ReadFile("pulumi-tpl.yaml")
	if err != nil {
		fmt.Println("Could not read file: ", err)
		return nil, err
	}

	var pulumiData map[string]interface{}
	err = yaml.Unmarshal(pulimiFile, &pulumiData)
	if err != nil {
       fmt.Println("Could not unmarshal the data: ", err)
	   return nil, err
	}

    pulumiData["name"] = suffixProjectName(projectName)
    
	//Access the config property 
	configProperty, ok := pulumiData["template"].(map[string]interface{})["config"]
	if !ok {
       fmt.Println("Could not find aws:region block")
	   return nil, fmt.Errorf("Could not find aws:region block")
	}

	configProperty.(map[string]interface{})["aws:region"].(map[string]interface{})["default"] = awsregion
	configProperty.(map[string]interface{})["pulumi:tags"].(map[string]interface{})["projectName"] = suffixProjectName(projectName)
	configProperty.(map[string]interface{})["pulumi:tags"].(map[string]interface{})["awsRegionDeployed"] = suffixProjectName(projectName)


	pulumiFileBytes, err := yaml.Marshal(pulumiData)
	if err != nil {
		fmt.Println("Could not Marshal the new data: ", err)
	}

	//Debugging purposes
	fmt.Println("Region: ", awsregion)
	fmt.Println(string(pulumiFileBytes))

	// Create a new pulumi.yaml file in the root directory
	err = os.WriteFile("../pulumi.yaml", pulumiFileBytes, 0644)
	if err != nil {
		fmt.Println("Could not create pulumi.yaml file: ", err)
		return nil, err
	}

	pulumiErr := runCli()
	if pulumiErr != nil {
		fmt.Println("Failed to create project: ", pulumiErr)
		return nil, pulumiErr
	}

	return pulumiFileBytes, nil

}

/*
  runCli: Runs the pulumi up command programatically

*/
func runCli() error {

	//check if pulumi is installed
	checkPulumiCmd := exec.Command("pulumi", "version")
	pulumiVersion, err := checkPulumiCmd.Output()

	if err != nil {
		fmt.Printf("Pulumi version %v is not installed", pulumiVersion)
		return err

		//TODO: Maybe install it

	} else {
		cmd := exec.Command("cd", projectRootDir)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Could not change directory", err)
			return err
		}
	
		pulumiUpCmd := exec.Command("pulumi", "up")
		pulumiUpErr := pulumiUpCmd.Run()
		if pulumiUpErr != nil {
			fmt.Println("Something went wrong", pulumiUpErr)
			return err
		}       
	}

	return nil

}

/*

This function adds a suffix to the provided project name to avoid duplicate names
Params: projectName
Returns: suffix-projectname

*/
func suffixProjectName( projectName string) string {

	rand.Seed( time.Now().UnixNano() )
	min := 100
	max := 10000

	return projectName + "-" + strconv.Itoa(rand.Intn( max - min + 1 ) ) 
}