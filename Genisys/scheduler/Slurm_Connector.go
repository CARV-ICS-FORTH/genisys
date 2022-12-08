// Copyright [2021] [FORTH-ICS]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"log"
	"math"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

var plot_data_init [][]string
var Plot_Data = make(map[string][][]string)

type ResourceUsage struct {
	CPU int
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

// Shellout_slurm() runs a bash squeue command on the Slurm host
func Shellout_slurm(command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, Shell_To_Use, "-c", command).Run(); err != nil {
		return err
	}
	return nil
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

// TODO extend Get_Pods_Slurm_Job_Id() to requests Slurm Jobs IDs across mutiple virtual clusters
// Get_Pods_Slurm_Job_Id() gets as input a deployments name and returns the
// Slurms's job ID corresponding to the specified deployment
func Get_Pods_Slurm_Job_Id(deployment_name string) string {

	Shellout_slurm("squeue --j=" + deployment_name + " -o\"%.7i %.9P %.8j %.8u %.2t %.10M %.6D %C %N\" | grep -v 'CPUS' > slurm_jobs/pod_job_id" + deployment_name)

	file, err := os.Open("slurm_jobs/pod_job_id" + deployment_name)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		return words[0]
	}
	file.Close()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""

}

// TODO extend Get_Pods_Slurm_Job_Id() to requests Slurm Jobs IDs across mutiple virtual clusters
// Get_Pods_Slurm_Job_Id() gets as input a deployments name and returns the
// Slurms's job ID corresponding to the specified deployment
func Get_Pods_Slurm_Job_Id_Per_Namespace(deployment_name string, namespace string, Virtual_Node_Name string) string {

	Shellout_slurm("./squeue.sh " + Virtual_Node_Name + " " + namespace + " " + deployment_name + " > slurm_jobs/pod_job_id" + deployment_name + namespace)

	file, err := os.Open("slurm_jobs/pod_job_id" + deployment_name + namespace)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		return words[0]
	}
	file.Close()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return ""

}

// Slurm_Job_Deleter() function iterates over the dummy deployments and
// checks if the corresponding Slurm Job exists,
// if the job does not exist it means that is either evicted, deleted or finished and the dummy Pods need to be removed
// if no Slurm Job is found Slurm_Job_Deleter() deletes the corresponding dummy Pods and frees the allocated resources
func Slurm_Job_Deleter() {

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			for _, namespace := range Namespaces {

				Pods, err := Get_Pods_Of_Namespace(namespace)
				if err != nil {
					panic(err)
				}

				for it := 0; it < len(Pods.Items); it++ {

					Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
					if err != nil {
						continue
					}

					Deployment_Name, _ := Get_Deployment_Name_Of_Pod(&Pods.Items[it])
					if err != nil {
						continue
					}

					//if Pods.Items[it].Status.Phase != "Running" && Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" {
					//	continue
					//}

					if Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
						continue
					}

					Slurm_Pods, _ := Get_Virtual_cluster_Pods(namespace)
					for _, Slurm_Pod := range Slurm_Pods.Items {
						slurm_job_exists := Get_Pods_Slurm_Job_Id_Per_Namespace(Deployment_Name, namespace, Slurm_Pod.ObjectMeta.Name)
						if slurm_job_exists == "" {
							Shellout_slurm("kubectl delete Deployment " + Deployment_Name + " -n " + namespace + " &")
						}
						break
					}

				}

				time.Sleep(100 * time.Millisecond)

			}
			time.Sleep(100 * time.Millisecond) //avoid spinning
		}
	}
}
