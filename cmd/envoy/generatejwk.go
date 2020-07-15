// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoy

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

//GenCmd to get org details
var GenCmd = &cobra.Command{
	Use:   "gen-jwk",
	Short: "Generate JSON Web Keys for Apigee Envoy Connector",
	Long:  "Generate JSON Web Keys for Apigee Envoy Connector",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err = Generatekeys(kid, folder); err != nil {
			return err
		}
		fmt.Println("Add the generated files to the Kubernetes secret:")
		fmt.Println("kubectl create secret generic {org}-{env}-policy-secret -n apigee --from-file=remote-service.key --from-file=remote-service.crt --from-file=remote-service.properties")
		return Generatekid(kid, folder)
	},
}

func init() {
	GenCmd.Flags().StringVarP(&folder, "folder", "f",
		"", "folder containing remote-service.* files")
	GenCmd.Flags().StringVarP(&kid, "kid", "k",
		time.Now().Format("2006-01-02T15:04:05"), "Key Identifier")
}
