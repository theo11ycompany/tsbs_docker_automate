// Exists only to make configs easier to feed into shell scripts
// Runs docker containers with resource in config
// Also cleans up after each run
// Also generated graphs on the data points as we want acc to the paper

// Pretty much a very useless orchestrator

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	// "sync"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var overall_x_axis_verticals map[string][]float64
var overall_y_axis_verticals map[string][]float64

func plot_non_flat_maps(x_axis map[string][]float64, y_axis map[string][]float64) {
	println("[INFO] Generating graphs for : ")
	for x := range x_axis {
		for y := range y_axis {
			p := plot.New()
			p.Title.Text = x + "/" + y
			p.X.Label.Text = x
			p.Y.Label.Text = y
			println("\t[INFO]", x, y)
			pts := make(plotter.XYs, len(x_axis[x]))
			for i := range x_axis[x] {
				pts[i].X = x_axis[x][i]
				pts[i].Y = y_axis[y][i]
			}

			plotutil.AddLinePoints(p, pts)
			if err := p.Save(8*vg.Inch, 8*vg.Inch, "./photos/"+x+"_"+strings.Replace(y, "/", "_", -1)+".png"); err != nil {
				panic(err)
			}
		}
	}

}

// No defining data models...takes time
// Equality between two maps, used to check is previous tsbs config is same as present to reduce time. This will probably cause too many problems lol
func map_equality(m1 map[string]interface{}, m2 map[string]interface{}) bool {
	if len(m1) == len(m2) {
		for k := range m1 {
			if m1[k].(string) == m2[k].(string) {
				continue
			} else {
				return false
			}
		}
		return true
	} else {
		return false
	}
}
func parse_load_results(stdout *bytes.Buffer, docker_cmd_opts map[string]interface{}, run_cmd_opts map[string]interface{}) {
	// println(docker_cmd_opts["cpus"].(string),overall_x_axis_verticals["cpus"])
	cpus, _ := strconv.ParseFloat(docker_cmd_opts["cpus"].(string), 32)
	memory, _ := strconv.ParseFloat(strings.Trim(docker_cmd_opts["memory"].(string), "m"), 32)
	workers, _ := strconv.ParseFloat(run_cmd_opts["workers"].(string), 32)
	overall_x_axis_verticals["cpus"] = append(overall_x_axis_verticals["cpus"], cpus)
	overall_x_axis_verticals["memory"] = append(overall_x_axis_verticals["memory"], memory)
	overall_x_axis_verticals["workers"] = append(overall_x_axis_verticals["workers"], workers)
	fin := false
	single_test_records := [][]float64{} // we'll use this later
	//find that one line thats empty
	for _, val := range strings.Split(stdout.String(), "\n") {
		line := strings.Trim(val, "\n")
		if !fin && line != "" {
			inner := []float64{}
			for _, val := range strings.Split(line, ",") {
				ival, _ := strconv.ParseFloat(val, 32)
				inner = append(inner, ival)
			}
			single_test_records = append(single_test_records, inner) //lets make use of this later
		} else if line == "" {
			fin = true
		} else {
			sp := strings.Split(line, "(")
			if len(sp) == 1 {
				continue
			} else {
				sp[1] = strings.Trim(sp[1], ")")
				sp2 := strings.Split(sp[1], " ")
				// println(sp2[3], sp2[2])
				ival, _ := strconv.ParseFloat(sp2[2], 32)
				overall_y_axis_verticals[sp2[3]] = append(overall_y_axis_verticals[sp2[3]], ival)
			}
		}
	}

}
func generate_cmd_from_tsbs_config(cmd_opts map[string]interface{}) {
	os.Remove(fmt.Sprintf("/tmp/influx-data.gz"))
	f, _ := os.Create(fmt.Sprintf("/tmp/influx-data.gz"))
	defer f.Close()
	cmd_opts_str := fmt.Sprintf("--use-case=devops --seed=%s --scale=%s --timestamp-start=%s --timestamp-end=%s --log-interval=%s --format=%s", cmd_opts["seed"], cmd_opts["scale"], cmd_opts["start_date"], cmd_opts["end_date"], cmd_opts["log_interval"], cmd_opts["target"])
	var stderr bytes.Buffer

	cmd_ctx := exec.Command("./tsbs_files/tsbs_generate_data", strings.Split(cmd_opts_str, " ")...)
	out, _ := cmd_ctx.StdoutPipe()
	gzw := gzip.NewWriter(f)
	defer gzw.Close()
	defer gzw.Flush()
	cmd_ctx.Stderr = &stderr
	fmt.Println("\t[CMD]", cmd_ctx.String())
	cmd_ctx.Start()
	io.Copy(gzw, out)
	err := cmd_ctx.Wait()
	if err != nil {
		fmt.Println("Error generating data", err, cmd_ctx.String())
		fmt.Println(stderr.String())
		panic(err)
	}

	os.Remove(fmt.Sprintf("/tmp/influx-query.gz"))
	f_q, _ := os.Create(fmt.Sprintf("/tmp/influx-query.gz"))
	defer f_q.Close()
	query_gen_cmd_opts := fmt.Sprintf("--use-case=devops --seed=%s --scale=%s --timestamp-start=%s --timestamp-end=%s --queries=%s --query-type=%s --format=%s", cmd_opts["seed"], cmd_opts["scale"], cmd_opts["start_date"], cmd_opts["end_date"], cmd_opts["queries"], cmd_opts["query_type"], cmd_opts["target"])

	cmd_ctx_q := exec.Command("./tsbs_files/tsbs_generate_queries", strings.Split(query_gen_cmd_opts, " ")...)
	out_q, _ := cmd_ctx_q.StdoutPipe()
	gzw_q := gzip.NewWriter(f_q)
	defer gzw_q.Close()
	defer gzw_q.Flush()
	stderr.Truncate(0)
	cmd_ctx_q.Stderr = &stderr
	fmt.Println("\t[CMD]", cmd_ctx_q.String())
	cmd_ctx_q.Start()
	io.Copy(gzw_q, out_q)
	err = cmd_ctx_q.Wait()
	if err != nil {
		fmt.Println("Error generating query", err, cmd_ctx_q.String())
		fmt.Println(stderr.String())
		panic(err)
	}
}

func load_cmd_from_tsbs_config(cmd_opts map[string]interface{}, target string) bytes.Buffer {

	load_cmd_opts := ""
	for k := range cmd_opts {
		if cmd_opts[k] == "" {
			continue
		}
		load_cmd_opts += "--" + k + "=" + cmd_opts[k].(string) + " "
	}

	for {
		var stderr bytes.Buffer
		var stdout bytes.Buffer
		cat_cmd := exec.Command("cat", fmt.Sprintf("/tmp/influx-data.gz"))
		gzip_cmd := exec.Command("gunzip")
		gzip_cmd.Stdin, _ = cat_cmd.StdoutPipe()
		cmd_ctx := exec.Command(fmt.Sprintf("./tsbs_files/tsbs_load_%s", target), strings.Split(load_cmd_opts, " ")...)
		cmd_ctx.Stdin, _ = gzip_cmd.StdoutPipe()
		cmd_ctx.Stdout = &stdout
		cmd_ctx.Stderr = &stderr

		cmd_ctx.Start()
		gzip_cmd.Start()
		cat_cmd.Start()

		fmt.Println("\t[CMD]", cmd_ctx.String())

		err := cmd_ctx.Wait()
		if err != nil {
			fmt.Println("Error running cat", err, cmd_ctx.String())
			fmt.Println(stderr.String())
			fmt.Println("\t[INFO] Ran into error in loading command, retryinh indefinetly \n",err)
			continue
		}
		println("here")
		err = gzip_cmd.Wait()
		if err != nil {
			fmt.Println("Error running cat", err, cmd_ctx.String())
			fmt.Println(stderr.String())
			fmt.Println("\t[INFO] Ran into error in loading command, retryinh indefinetly \n",err)
			continue
		}
		println("here")
		err = cat_cmd.Wait()
		if err != nil {
			fmt.Println("Error running cat", err, cmd_ctx.String())
			fmt.Println(stderr.String())
			fmt.Println("\t[INFO] Ran into error in loading command, retryinh indefinetly \n",err)
			continue
		}
		println("here")
		return stdout
	}

	// wg := sync.WaitGroup{}

	// wg.Add(1)
	// go func() {
	// 	gzip_cmd.Start()
	// 	defer wg.Done()
	// 	err := gzip_cmd.Wait()
	// 	if err != nil {
	// 		fmt.Println("Error running gzip", err, cmd_ctx.String())
	// 		fmt.Println(stderr.String())
	// 		panic(err)
	// 	}
	// }()
	//
	// time.Sleep(2 * time.Second)
	//
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	cmd_ctx.Start()
	// 	err := cmd_ctx.Wait()
	// 	if err != nil {
	// 		fmt.Println("Error loading", err, cmd_ctx.String())
	// 		fmt.Println(stderr.String())
	// 		panic(err)
	// 	}
	// }()
	//
	// time.Sleep(2 * time.Second)
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	cat_cmd.Start()
	// 	err := cat_cmd.Wait()
	// 	if err != nil {
	// 		fmt.Println("Error running cat", err, cmd_ctx.String())
	// 		fmt.Println(stderr.String())
	// 		panic(err)
	// 	}
	// }()
	// wg.Wait()
	//
}

func DockerBuildAndRun(dockerfile_path string, docker_image_name string, docker_config map[string]interface{}) string {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := fmt.Sprintf("build %s -t %s", dockerfile_path, docker_image_name)
	cmd_ctx := exec.Command("docker", strings.Split(cmd, " ")...)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
	fmt.Println("\t[CMD]", cmd_ctx.String())
	err := cmd_ctx.Run()
	if err != nil {
		fmt.Println("[ERROR] Erorr building image", err, cmd)
		fmt.Println("[ERROR]", stderr.String())
		panic(err)
	}

	cmd = fmt.Sprintf("run -d --name questdb -m %s --cpus=%s -p 9009:9009 questdb/questdb:6.0.4", docker_config["memory"], docker_config["cpus"]) //[TODO] lets this port be configurable
	cmd_ctx = exec.Command("docker", strings.Split(cmd, " ")...)
	fmt.Println("\t[CMD]", cmd_ctx.String())
	out.Truncate(0)
	stderr.Truncate(0)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
	err = cmd_ctx.Run()
	if err != nil {
		fmt.Println("[ERROR]", stderr.String())
		fmt.Println("[ERROR] Error running images, check docker daemon and port availability...")
		panic(err)
	}
	fmt.Println("\t[INFO] Created container, id : ", out.String())
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
		fmt.Println("[ERROR] Error stopping container...error stack : ")
		fmt.Println("[ERROR]", stderr.String())
		panic(err)
	}

	cmd = fmt.Sprintf("rm %s", container_id[:5]) // for some reason fully qualified container id is causing missing pages in docker daemon?!?
	cmd_ctx = exec.Command("docker", strings.Split(cmd, " ")...)
	cmd_ctx.Stdout = &out
	cmd_ctx.Stderr = &stderr
	err = cmd_ctx.Run()
	if err != nil {
		fmt.Println("[ERROR] Error removing container...error stack : ")
		fmt.Println("[ERROR]", stderr.String())
		panic(err)
	}
}

func main() {
	fmt.Println("[NOTE] This orchestrator needs sudoless docker!!")
	fmt.Println("[NOTE] Every set has a fixed database constraint")
	fmt.Println("[INFO] Parsing config file...")
	overall_y_axis_verticals = make(map[string][]float64)
	overall_x_axis_verticals = make(map[string][]float64)
	overall_x_axis_verticals["cpus"] = []float64{}
	overall_x_axis_verticals["memory"] = []float64{}
	overall_x_axis_verticals["workers"] = []float64{}
	overall_y_axis_verticals["metrics/sec"] = []float64{}
	overall_y_axis_verticals["rows/sec"] = []float64{}

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
			fmt.Println("[INFO] running test set ", i+1, "...")
			test_obj := test_interface.(map[string]interface{})
			dockerfile_path := test_obj["docker_file"].(string)
			docker_image_name := test_obj["docker_image_name"].(string)
			docker_config := test_obj["docker_config"].(map[string]interface{})
			fmt.Println("\t[INFO] Starting docker containers with constraints...")
			container_id := DockerBuildAndRun(dockerfile_path, docker_image_name, docker_config)
			println("\t[INFO]Sleeping for 8 seconds between test runs")
			time.Sleep(8 * time.Second)
			tsbs_gen_config := test_obj["tsbs_gen_config"].(map[string]interface{})
			if prev_gen_config != nil && map_equality(prev_gen_config, tsbs_gen_config) {
				fmt.Println("\t[INFO] Reusing previously generated data...")
			} else {
				fmt.Println("\t[INFO] Generating data and queries...")
				generate_cmd_from_tsbs_config(tsbs_gen_config)
			}
			target := test_obj["target"].(string)
			tsbs_run_config := test_obj["tsbs_run_config"].(map[string]interface{})
			stdout_buffer := load_cmd_from_tsbs_config(tsbs_run_config, target)
			parse_load_results(&stdout_buffer, docker_config, tsbs_run_config)
			fmt.Println("\t[INFO] Stopping docker containers...")
			DockerStopContainer(container_id)
			time.Sleep(5 * time.Second)
			// parse_test(docker_config, tsbs_run_config)
			prev_gen_config = tsbs_gen_config

			plot_non_flat_maps(overall_x_axis_verticals, overall_y_axis_verticals)
		}
	}

}

func parse_test(docker_cmd_opts map[string]interface{}, run_cmd_opts map[string]interface{}) {
	output := `time,per. metric/s,metric total,overall metric/s,per. row/s,row total,overall row/s
1699973783,1548249.47,1.548280E+07,1548249.47,137997.28,1.380000E+06,137997.28
1699973793,1390843.33,2.939240E+07,1469543.87,123989.60,2.620000E+06,130993.21
1699973803,1402564.97,4.341800E+07,1447218.39,125000.44,3.870000E+06,128995.70
1699973813,1359159.27,5.700840E+07,1425205.61,121010.62,5.080000E+06,126999.61
1699973823,1278397.62,6.979240E+07,1395844.04,113999.79,6.220000E+06,124399.65
1699973833,1335973.64,8.315280E+07,1385865.25,118994.09,7.410000E+06,123498.69
1699973843,1301441.86,9.616760E+07,1373804.58,115996.60,8.570000E+06,122426.94
1699973853,1268276.55,1.088496E+08,1360614.42,113006.82,9.700000E+06,121249.50
1699973863,1279850.58,1.216484E+08,1351640.51,113997.38,1.084000E+07,120443.70
1699973873,1357513.34,1.352240E+08,1352227.81,120995.84,1.205000E+07,120498.91
1699973883,1278904.58,1.480132E+08,1345562.04,113998.63,1.319000E+07,119907.98
1699973893,1301543.59,1.610280E+08,1341894.03,116005.67,1.435000E+07,119582.80
1699973903,1269184.81,1.737200E+08,1336300.98,112998.65,1.548000E+07,119076.32
1699973913,1233743.97,1.860576E+08,1328975.42,109998.57,1.658000E+07,118427.91
1699973923,1223574.88,1.982928E+08,1321949.05,109004.89,1.767000E+07,117799.74
1699973933,1200359.22,2.102976E+08,1314348.99,106989.23,1.874000E+07,117124.02
1699973943,1190380.74,2.222000E+08,1307057.60,106012.53,1.980000E+07,116470.48

Summary:
loaded 232704000 metrics in 177.829sec with 10 workers (mean rate 1308584.92 metrics/sec)
loaded 20736000 rows in 177.829sec with 10 workers (mean rate 116606.58 rows/sec)`
	mocking_stdout := []byte(output)
	parse_load_results(bytes.NewBuffer(mocking_stdout), docker_cmd_opts, run_cmd_opts)
}

