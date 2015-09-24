package main

import (
	"fmt"
	"github.com/codegangsta/martini-contrib/binding"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"os"
	"os/exec"
	"text/template"
)

type DockerInfo struct {
	BRANCH     string `json:"git_branch" binding:"required"`
	HASH       string `json:"git_hash" binding:"required"`
	REPO_OWNER string `json:"repo_owner" binding:"required"`
	REMOTE_LOC string `json:"path" binding:"required"`
	REPO_NAME  string `json:"repo_name" binding:"required"`
}

func buildDockerContainer() {
	// We can pass in a callback here, or just handle the status update
	// request from this function
	buildCommand := exec.Command("docker", "build", "--no-cache=True", "--tags='franklin_builder_tmp:tmp'", ".")
	if err := buildCommand.Run(); err != nil {
		fmt.Println(os.Stderr, err)
	}

	tearDown := exec.Command("scripts/tear_down_project.sh")
	if err := tearDown.Run(); err != nil {
		fmt.Println(os.Stderr, err)
	}

	os.Remove("tmp/")
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Post("/build", binding.Bind(DockerInfo{}), BuildDockerFile)
	m.Run()

}

func BuildDockerFile(p martini.Params, r render.Render, dockerInfo DockerInfo) {
	tmp_dir := "tmp"

	// Create a new Dockerfile template parses template definition
	docker_tmpl, err := template.ParseFiles("templates/dockerfile.tmplt")
	HandleErr(err)

	// Create tmp directory
	err = os.Mkdir(tmp_dir, 0770)
	HandleErr(err)

	// Create file
	f, err := os.Create(tmp_dir + "/Dockerfile")
	HandleErr(err)
	defer f.Close()

	//Apply the Dockerfile template to the docker info from the request
	err = docker_tmpl.Execute(f, dockerInfo)
	HandleErr(err)

	// Build the docker container.
	go buildDockerContainer()
	r.JSON(200, map[string]interface{}{"success": true})

}