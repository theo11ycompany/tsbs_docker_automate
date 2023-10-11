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

// No defining data models...takes time
func map_equality(m1 map[string]interface{}, m2 map[string]interface{}) bool{
   if(len(m1) == len(m2)){
      for k := range m1{
         if m1[k].(string) == m2[k].(string){
            continue
         }else{
            return false
         }
      }
      return true
   }else{
      return false
   }
}
func generate_cmd_from_tsbs_config(cmd_opts map[string]interface{}, script string) {
   cmd_opts_str := ""
	var out bytes.Buffer
	var stderr bytes.Buffer
	for key, value := range cmd_opts {
      if value.(string)!=""{
         cmd_opts_str += "--"+key+"="+value.(string)+" "
      }
	}

   cmd_ctx := exec.Command(script, strings.Split(cmd_opts_str," ")...)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
   fmt.Println(cmd_ctx.String())
	cmd_ctx.Start()
   err := cmd_ctx.Wait()
	if err != nil {
		fmt.Println("Error generating data and queries...", err, cmd_ctx.String())
		fmt.Println(stderr.String())
		panic(err)
	}
}

func run_cmd_from_tsbs_config(cmd_opts map[string]interface{}, load_script string, run_script string){

}

func DockerBuildAndRun(dockerfile_path string, docker_image_name string, docker_config map[string]interface{}) string {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := fmt.Sprintf("build %s -t %s", dockerfile_path, docker_image_name)
	cmd_ctx := exec.Command("docker", strings.Split(cmd, " ")...)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
	err := cmd_ctx.Run()
	if err != nil {
		fmt.Println("Erorr building image", err, cmd)
		fmt.Println(stderr.String())
		panic(err)
	}

	cmd = fmt.Sprintf("run -d -m %s --cpus=%s -p 9000:9000 %s:latest", docker_config["memory"], docker_config["cpus"], docker_image_name)
	cmd_ctx = exec.Command("docker", strings.Split(cmd, " ")...)
	out.Truncate(0)
	stderr.Truncate(0)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
	err = cmd_ctx.Run()
	if err != nil {
		fmt.Println(stderr.String())
		fmt.Println("Error running images, check docker daemon and port availability...")
		panic(err)
	}

	fmt.Println("\tCreated container, id : ", out.String())
	return out.String()

}
func DockerStopContainer(container_id string) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := fmt.Sprintf("stop %s", container_id[:5]) // for some reason fully qualified container id is causing missing pages in docker daemon?!?
	cmd_ctx := exec.Command("docker", strings.Split(cmd, " ")...)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
	err := cmd_ctx.Run()
	if err != nil {
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
      prev_gen_config := make(map[string]interface{})
		for i, test_interface := range test {
			fmt.Println("running test set ", i+1, "...")
			test_obj := test_interface.(map[string]interface{})
			dockerfile_path := test_obj["docker_file"].(string)
			docker_image_name := test_obj["docker_image_name"].(string)
			docker_config := test_obj["docker_config"].(map[string]interface{})
			// test_shell_file := test_obj["test_shell_file_path"]
			fmt.Println("\tStarting docker containers with constraints...")
			container_id := DockerBuildAndRun(dockerfile_path, docker_image_name, docker_config)
			tsbs_gen_config := test_obj["tsbs_gen_config"].(map[string]interface{})
			tsbs_gen_script := test_obj["gen_script"].(string)
         if prev_gen_config!=nil && map_equality(prev_gen_config, tsbs_gen_config){
			   fmt.Println("\tReusing previously generated data...")
         }else{
			   fmt.Println("\tGenerating data and queries...")
			   generate_cmd_from_tsbs_config(tsbs_gen_config, tsbs_gen_script)
         }
         tsbs_query_run_script := test_obj["query_run_script"].(string)
         tsbs_load_script := test_obj["load_script"].(string)
			tsbs_run_config := test_obj["tsbs_run_config"].(map[string]interface{})
         run_cmd_from_tsbs_config(tsbs_run_config, tsbs_load_script, tsbs_query_run_script)
			fmt.Print("\tStopping docker containers...")
			DockerStopContainer(container_id)
         prev_gen_config = tsbs_gen_config
		}
	}

}
