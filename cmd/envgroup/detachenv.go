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

package envgroup

import (
	"internal/apiclient"

	"internal/client/envgroups"

	"github.com/spf13/cobra"
)

// DetachCmd to get env group
var DetachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach an env from an Environment Group",
	Long:  "Detach an env from an Environment Group",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		return apiclient.SetApigeeOrg(org)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		_, err = envgroups.DetachEnvironment(name, environment)
		return
	},
}

func init() {
	DetachCmd.Flags().StringVarP(&org, "org", "o",
		"", "Apigee organization name")

	DetachCmd.Flags().StringVarP(&name, "name", "n",
		"", "Name of the environment group")

	DetachCmd.Flags().StringVarP(&environment, "env", "e",
		"", "Name of the environment")

	_ = DetachCmd.MarkFlagRequired("name")
	_ = DetachCmd.MarkFlagRequired("env")
}
