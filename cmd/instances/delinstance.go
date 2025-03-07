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

package instances

import (
	"internal/apiclient"

	"internal/client/instances"

	"github.com/spf13/cobra"
)

// DeleteCmd to get instance
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an Instance",
	Long:  "Delete an Instance",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		return apiclient.SetApigeeOrg(org)
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		_, err = instances.Delete(name)
		return
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&org, "org", "o",
		"", "Apigee organization name")

	DeleteCmd.Flags().StringVarP(&name, "name", "n",
		"", "Name of the instance")

	_ = DeleteCmd.MarkFlagRequired("name")
}
