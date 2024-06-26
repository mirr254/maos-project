package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"maos-cloud-project-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/sirupsen/logrus"
)

type Project struct {
	ProjectName string `json:"project_name"`
	AwsRegion   string `json:"aws_region"`
	StackName   string `json:"environment"`
}

type ProjectResponse struct {
	StackName  string `json:"environment"`
	URL 	   string `json:"url"`
	// Status      string `json:"status"`
}

/*
CreateProject: function is responsible for creating a new project on pulumi dashboard.
A Client can have multiple projects.

*/

func CreateProject(c *gin.Context){

	// rootDir, err := utils.GetRootDir()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get root directory"})
	// 	return
	// }

	var project Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	suffixedProjectName := suffixProjectName(project.ProjectName)

	projectConfig := utils.BuildProjectConfig(project.ProjectName, project.AwsRegion)

	// Generate Pulumi.yaml file dynamically
	err := utils.GeneratePulumiYAML(projectConfig, "Pulumi.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Println("Could not generate Pulumi.yaml file: ", err)
		return
	}

	ctx := context.Background()

	stack := project.StackName
	s, err := auto.NewStackInlineSource(ctx, stack, suffixedProjectName, PulumiProgram, auto.Program(PulumiProgram) )
	if err != nil {
		if auto.IsCreateStack409Error(err) {
			logrus.Error("Stack Exists error: ", err)
			// c.JSON(http.StatusConflict, gin.H{"error": "Stack already exists"})
			// return
		} else {
			logrus.Error("Stack Error ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create stack"})
			return
		}

	}


	// Set stack configuration
	projectConfigMap := utils.ConvertProjectConfigToAutoConfigMap(projectConfig)
	s.SetAllConfig(ctx, projectConfigMap)


	//deploy stack
	upRes, err := s.Up(ctx, optup.ProgressStreams(os.Stdout))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not deploy stack"})
		logrus.Error("Could not deploy stack", err)
		return
	}

	// Convert output to json and print
	outputJson, err := json.Marshal(upRes.Outputs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not marshal output"})
		return
	}

	fmt.Println(string(outputJson))

	c.JSON(http.StatusCreated, string(outputJson))
	return

}

func PulumiProgram(ctx *pulumi.Context) error {
	// Implement your Pulumi program here
	ctx.Export("Success!", pulumi.Sprintf("success"))
	return nil
}


/*
This function adds a suffix to the provided project name to avoid duplicate names
Params: projectName
Returns: suffix-projectname
*/
func suffixProjectName(projectName string) string {

	rand.Seed(time.Now().UnixNano())
	min := 100
	max := 10000

	fmt.Sprintf("Project name is %s", projectName+"-"+strconv.Itoa(rand.Intn(max-min+1)))

	return projectName + "-" + strconv.Itoa(rand.Intn(max-min+1))
}
