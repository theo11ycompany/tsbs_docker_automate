// Exists only to make configs easier to feed into shell scripts
// Runs docker containers with resource in config
// Also cleans up after each run
// Also generated graphs on the data points as we want acc to the paper

// Pretty much a very useless orchestrator

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// type ConfigObj Test{
//
// }
//
// type ConfigObj struct {
//    Tests []Test `json:"tests"`
// }

func run_tsbs_script()

func DockerBuildAndRun(dockerfile_path string, docker_image_name string, docker_config map[string]interface{}) string{
   var out bytes.Buffer
   var stderr bytes.Buffer

   cmd := fmt.Sprintf("build %s -t %s", dockerfile_path, docker_image_name)
   cmd_ctx:= exec.Command("docker", strings.Split(cmd, " ")...)
   cmd_ctx.Stdout = &out
   cmd_ctx.Stderr = &stderr
   err := cmd_ctx.Run()
   if err!=nil{
      fmt.Println("Erorr building image", err, cmd)
      fmt.Println(stderr.String())
      panic(err)
   }

   cmd = fmt.Sprintf("run -d -m %s --cpus=%s -p 9000:9000 %s:latest", docker_config["memory"], docker_config["cpus"], docker_image_name)
   cmd_ctx = exec.Command("docker",strings.Split(cmd," ")...)
   out.Truncate(0)
   stderr.Truncate(0)
   cmd_ctx.Stdout = &out
   cmd_ctx.Stderr = &stderr
   err = cmd_ctx.Run()
   if err!=nil{
      fmt.Println(stderr.String())
      fmt.Println("Error running images, check docker daemon and port availability...");
      panic(err)
   }

   fmt.Println("\tCreated container, id : ", out.String())
   return out.String();

}
func DockerStopContainer(container_id string) {
   var out bytes.Buffer
   var stderr bytes.Buffer
   cmd := fmt.Sprintf("stop %s", container_id[:5]); // for some reason fully qualified container id is causing missing pages in docker daemon?!?
   cmd_ctx := exec.Command("docker", strings.Split(cmd, " ")...)
   cmd_ctx.Stdout = &out
   cmd_ctx.Stderr = &stderr
   err := cmd_ctx.Run()
   if err!=nil{
      fmt.Println("Error stopping container...error stack : ")
      fmt.Println(stderr.String())
      panic(err)
   }
}

func main() {
   fmt.Println("NOTE : This orchestrator needs sudoless docker!!")
   fmt.Println("NOTE : Every set has a fixed database constraint")
	fmt.Println("Parsing config file...")
	dat, err := os.ReadFile("./config.json")
	if err != nil {
		println("err : ", err)
	}
	var dat_obj map[string]interface{}
	err = json.Unmarshal(dat, &dat_obj)
	if err != nil {
		println("err : ", err)
	} else {
		test := dat_obj["tests"].([]interface{})
      fmt.Println("")
      fmt.Println("")
		for i, test_interface := range test {
			fmt.Println("running test set ", i+1, "...")
			test_obj := test_interface.(map[string]interface{})
			dockerfile_path := test_obj["docker_file"].(string)
			docker_image_name := test_obj["docker_image_name"].(string)
			docker_config := test_obj["docker_config"].(map[string]interface{})
			// test_shell_file := test_obj["test_shell_file_path"]
	      fmt.Println("\tStarting docker containers with constraints...")
         container_id := DockerBuildAndRun(dockerfile_path, docker_image_name, docker_config)
         tsbs_config_list := test_obj["tsbs_configs"].([]interface{})
         for j, tsbs_opts_objs := range tsbs_config_list{
            fmt.Println("\tRunning TSBS loads on database...")
            fmt.Println("\tRunning load",j,":",tsbs_opts_objs)
         }
	      fmt.Print("\tStopping docker containers...")
         DockerStopContainer(container_id)
		}
	}

}
