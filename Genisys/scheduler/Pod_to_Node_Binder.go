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
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"context"
)

// Bind_Pod_To_Node() takes as input a Pod and the a Node's Name as a String
// and place the Pod to the specified Node.
// Used by the Pod scheduler.

func Bind_Pod_To_Node(Pod v1.Pod, Node string) {

	binding := &v1.Binding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Binding",
		},
		ObjectMeta: Pod.ObjectMeta,
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       Node,
		},
	}

	Client_Set, err1 := kubernetes.NewForConfig(Config)
	if err1 != nil {
		panic(err1.Error())
	}
	Client_Set.CoreV1().Pods(Pod.ObjectMeta.Namespace).Bind(context.Background(), binding, metav1.CreateOptions{})
	message := fmt.Sprintf("Successfully assigned %s to %s", Pod.ObjectMeta.Name, Node)
	log.Println(message)
}

func Bind_Pod_To_Node_Controller(Pod *v1.Pod, Node v1.Node) error {
	Bind_Pod_To_Node(*Pod, Node.ObjectMeta.Name)
	return nil
}
