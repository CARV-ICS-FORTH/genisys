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
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v2 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const ShellToUse = "/bin/sh"

func Shellout(command string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, ShellToUse, "-c", command).Run(); err != nil {

	}
}

func main() {

	var kubeconfig *string
	var deployment_name string
	var num_of_nodes int64
	var total_cpus_per_node int64

	if os.Args[1] == "help" {
		fmt.Println("ARGS = Deploymemt Name -- Number of Replicas -- Total CPUs per Node")
		return
	}

	deployment_name = os.Args[1]
	num_of_nodes, _ = strconv.ParseInt(os.Args[2], 10, 64)
	total_cpus_per_node, _ = strconv.ParseInt(os.Args[3], 10, 64)

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	namespace_bytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	namespace := string(namespace_bytes)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": deployment_name,
				"labels": map[string]interface{}{
					"app": "Slurm",
				},
				"annotations": map[string]interface{}{
					"flag":                        "0",
					"cpu-bandwidth":               strconv.FormatInt(total_cpus_per_node*1000, 10),
					"app":                         "MPI-JOB",
					"type":                        "MPI-JOB",
					"dynamic-resource-management": "1",
				},
			},
			"spec": map[string]interface{}{
				"replicas": num_of_nodes,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "Slurm",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							//"k8s.v1.cni.cncf.io/networks": "macvlan-conf",
							"app":           "MPI-JOB",
							"schedulerName": "genisys",
						},
						"labels": map[string]interface{}{
							"app": "Slurm",
						},
					},
					"spec": map[string]interface{}{
						"schedulerName": "genisys",
						"containers": []map[string]interface{}{
							{
								"name":  "mofedtestdeployment",
								"image": "wardsco/sleep:latest",
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created deployment: " + result.GetName())

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	ready := 1
	for ready < 2 {

		//check if deployment is ready every 1 second
		time.Sleep(1 * time.Second)

		deployments := clientset.AppsV1().Deployments(namespace)
		deployment_k8s, err := deployments.Get(context.Background(), deployment_name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		if deployment_k8s.Status.ReadyReplicas == deployment_k8s.Status.Replicas && deployment_k8s.Status.Replicas != 0 {
			fmt.Println("Scheduled deployment: " + result.GetName())
			fmt.Println(deployment_k8s.Status.Replicas)
			fmt.Println(deployment_k8s.Status.ReadyReplicas)
			config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
			if err != nil {
				config, err = rest.InClusterConfig()
				if err != nil {
					continue
				}
			}
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				continue
			}

			//get the cluster's pod list
			pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), v2.ListOptions{})
			if err != nil {
				continue
			}
			nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
				LabelSelector: "mpi=1",
			})
			if err != nil {
				continue
			}

			Nodes_K8s_to_Slurm_Map := make(map[int]string)
			for _, node := range nodes.Items {
				key_int, _ := strconv.Atoi(strings.Replace(strings.Replace(node.Spec.PodCIDR, ".", "", -1), "/", "", -1))
				Nodes_K8s_to_Slurm_Map[key_int] = node.Name
			}

			Nodes_K8s_to_Slurm_Map_Ordered_Keys := make([]int, 0, len(Nodes_K8s_to_Slurm_Map))
			for key := range Nodes_K8s_to_Slurm_Map {
				Nodes_K8s_to_Slurm_Map_Ordered_Keys = append(Nodes_K8s_to_Slurm_Map_Ordered_Keys, key)
			}

			sort.Ints(Nodes_K8s_to_Slurm_Map_Ordered_Keys)

			for _, podItem := range pods.Items {
				deploymentname := strings.Split(podItem.ObjectMeta.Name, "-")
				if deploymentname[0] == deployment_name {
					for Slurm_Node_Index, key := range Nodes_K8s_to_Slurm_Map_Ordered_Keys {
						if Nodes_K8s_to_Slurm_Map[key] == podItem.Spec.NodeName {
							f, err := os.OpenFile("/tmp/slurm_nums/"+deploymentname[0], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
							if err != nil {
								log.Fatal(err)
							}
							Slurm_Node_Index_String := strconv.Itoa(Slurm_Node_Index)
							if _, err = f.WriteString(Slurm_Node_Index_String + "\n"); err != nil {
								log.Fatal(err)
							}
							fmt.Println("Slurm file is ready for deployment: " + result.GetName())
							defer f.Close()

						}
					}
				}
			}
			return
		}

	}
}
