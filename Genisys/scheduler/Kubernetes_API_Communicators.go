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
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v2 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
)

func Get_Dummy_MPI_Pods(Namespace_Input string) (*v1.PodList, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Pods, err := Client_Set.CoreV1().Pods(Namespace_Input).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + "Slurm"})
	if err != nil {
		panic(err.Error())
	}

	return Pods, nil
}

func Get_Dummy_MPI_Pods_All_Namespaces() (*v1.PodList, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Pods, err := Client_Set.CoreV1().Pods(Global_Namespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + "Slurm"})
	if err != nil {
		panic(err.Error())
	}

	return Pods, nil
}

func Get_Virtual_cluster_Pods(Namespace_Input string) (*v1.PodList, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Pods, err := Client_Set.CoreV1().Pods(Namespace_Input).List(context.Background(), metav1.ListOptions{LabelSelector: "app=" + "mpi"})
	if err != nil {
		panic(err.Error())
	}

	return Pods, nil
}

func Get_Pod_By_Name(Pod_Name string) (*v1.Pod, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Pods, err := Client_Set.CoreV1().Pods(Global_Namespace).List(context.Background(), metav1.ListOptions{FieldSelector: "spec.schedulerName=" + *Scheduler_Name})
	if err != nil {
		panic(err.Error())
	}

	for it := 0; it < len(Pods.Items); it++ {
		if Pods.Items[it].ObjectMeta.Name == Pod_Name {
			return &Pods.Items[it], nil
		}
	}

	return nil, nil
}

func Get_Deployment_Of_Pod(Pod *v1.Pod) (*appsv1.Deployment, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}
	Deployments := Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
	Deployment_Name := strings.Split(Pod.ObjectMeta.Name, "-")
	Deployment, err := Deployments.Get(context.Background(), Deployment_Name[0], metav1.GetOptions{})

	return Deployment, err
}

func Get_Deployment_Name_Of_Pod(Pod *v1.Pod) (string, error) {

	Deployment_Name := strings.Split(Pod.ObjectMeta.Name, "-")
	_, err := Get_Deployment_Of_Pod(Pod)
	if err != nil {
		return Deployment_Name[0], err
	}
	return Deployment_Name[0], err
}

func Get_Pod_Namespace(Pod_Name string) string {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	for _, namespace := range Namespaces {

		Pods, err := Client_Set.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{FieldSelector: "spec.schedulerName=" + *Scheduler_Name})
		if err != nil {
			panic(err.Error())
		}

		for it := 0; it < len(Pods.Items); it++ {
			if Pods.Items[it].ObjectMeta.Name == Pod_Name {
				return namespace
			}
		}
	}
	return ""
}

func Get_Pods_all_Namespaces() (*v1.PodList, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Pods, err := Client_Set.CoreV1().Pods(Global_Namespace).List(context.Background(), metav1.ListOptions{FieldSelector: "spec.schedulerName=" + *Scheduler_Name})
	if err != nil {
		panic(err.Error())
	}
	return Pods, nil
}

func Get_Pods_Of_Namespace(namespace string) (*v1.PodList, error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Pods, err := Client_Set.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{FieldSelector: "spec.schedulerName=" + *Scheduler_Name})
	if err != nil {
		panic(err.Error())
	}
	return Pods, nil
}

func Get_All_Nodes() (*v1.NodeList, error) {

	var err error
	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Nodes, err := Client_Set.CoreV1().Nodes().List(context.Background(), v2.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return Nodes, nil
}

func Node_Capacity() (string, error) {

	var err error

	Node_List, err := Get_All_Nodes()
	if err != nil {
		panic(err.Error())
	}

	return Node_List.Items[0].Status.Capacity.Cpu().String(), nil
}

func Node_Capacity_Of_Node(Node_Name string) int64 {

	var err error
	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Node, err := Client_Set.CoreV1().Nodes().Get(context.Background(), Node_Name, v2.GetOptions{})
	if err != nil {
		return -1
	}

	CPU_Node_Capacity_Float, _ := strconv.ParseFloat(Node.Status.Capacity.Cpu().String(), 64)
	Node_Capacity := int64(CPU_Node_Capacity_Float * 1000 * Max_Node_Allocation_Percentage)
	return Node_Capacity
}

func Get_MPI_Nodes() (*v1.NodeList, error) {

	var err error
	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}
	Nodes, err := Client_Set.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: "mpi=1",
	})
	if err != nil {
		panic(err.Error())
	}
	return Nodes, nil
}
