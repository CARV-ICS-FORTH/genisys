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
	"encoding/csv"
	"log"
	"os"
)

func Per_App_Log_File_Reset() {

	if _, err := os.Stat("plots/"); os.IsNotExist(err) {
		err := os.Mkdir("plots", 0777)
		if err != nil {
			panic(err)
		}
	}

	Pods, err := Get_Pods_all_Namespaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, podItem := range Pods.Items {

		Deployment_Name, err := Get_Deployment_Name_Of_Pod(&podItem)
		if err != nil {
			continue
		}
		err = os.Truncate("plots/"+Deployment_Name+".csv", 0)
		if err != nil {
			continue
		}
		err = os.Chmod("plots/"+Deployment_Name+".csv", 0777)
		if err != nil {
			continue
		}

		f, err := os.OpenFile("plots/"+Deployment_Name+".csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
		if err != nil {
			continue
		}
		defer f.Close()

		Plot_Data["plots/"+Deployment_Name+".csv"] = append(Plot_Data["plots/"+Deployment_Name+".csv"], []string{"time_x", "current_performance", "target_performance", "cpu_node_percentage", "net_node_percentage", "io_node_percentage"})

		w := csv.NewWriter(f)
		w.WriteAll(Plot_Data["plots/"+Deployment_Name+".csv"])

		if err := w.Error(); err != nil {
			continue
		}
	}
}

func Per_App_Log_File_Creator() {

	Pods, err := Get_Pods_all_Namespaces()
	if err != nil {
		return
	}
	for _, podItem := range Pods.Items {
		Deployment_Name, err := Get_Deployment_Name_Of_Pod(&podItem)
		if err != nil {
			continue
		}
		err = os.Chmod("plots/"+Deployment_Name+".csv", 0777)
		if err != nil {
			continue
		}
		file, err := os.OpenFile("plots/"+Deployment_Name+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			continue
		}
		defer file.Close()
	}

}

func Lines_In_File_Counter(fileName string) int {
	f, _ := os.Open(fileName)
	// Create new Scanner.
	scanner := bufio.NewScanner(f)
	result := []string{}
	// Use Scan.
	for scanner.Scan() {
		line := scanner.Text()
		// Append line to result.
		result = append(result, line)
	}
	return len(result)
}
