// Copyright (c) 2016-2020 Tigera, Inc. All rights reserved.

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

package commands

import (
	"fmt"
	"strings"

	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/calicoctl/calicoctl/commands/common"
	"github.com/projectcalico/calicoctl/calicoctl/commands/constants"
)

func Create(args []string) error {
	doc := constants.DatastoreIntro + `Usage:
  calicoctl create --filename=<FILENAME> [--skip-exists] [--config=<CONFIG>] [--namespace=<NS>] [--dry-run]

Examples:
  # Create a policy using the data in policy.yaml.
  calicoctl create -f ./policy.yaml

  # Create a policy based on the JSON passed into stdin.
  cat policy.json | calicoctl create -f -

Options:
  -h --help                 Show this screen.
  -f --filename=<FILENAME>  Filename to use to create the resource.  If set to
                            "-" loads from stdin.
     --skip-exists          Skip over and treat as successful any attempts to
                            create an entry that already exists.
  -c --config=<CONFIG>      Path to the file containing connection
                            configuration in YAML or JSON format.
                            [default: ` + constants.DefaultConfigPath + `]
  -n --namespace=<NS>       Namespace of the resource.
                            Only applicable to NetworkPolicy, NetworkSet, and WorkloadEndpoint.
                            Uses the default namespace if not specified.
  -d --dry-run              Dry run of calicoctl create.
                            Checks the validity and syntax of policies before applying.

Description:
  The create command is used to create a set of resources by filename or stdin.
  JSON and YAML formats are accepted.

  Valid resource types are:

    * bgpConfiguration
    * bgpPeer
    * felixConfiguration
    * globalNetworkPolicy
    * globalNetworkSet
    * hostEndpoint
    * ipPool
    * kubeControllersConfiguration
    * networkPolicy
    * networkSet
    * node
    * profile
    * workloadEndpoint

  Attempting to create a resource that already exists is treated as a
  terminating error unless the --skip-exists flag is set.  If this flag is set,
  resources that already exist are skipped.

  The output of the command indicates how many resources were successfully
  created, and the error reason if an error occurred.  If the --skip-exists
  flag is set then skipped resources are included in the success count.

  The resources are created in the order they are specified.  In the event of a
  failure creating a specific resource it is possible to work out which
  resource failed based on the number of resources successfully created.
`
	parsedArgs, err := docopt.Parse(doc, args, true, "", false, false)
	if err != nil {
		return fmt.Errorf("Invalid option: 'calicoctl %s'. Use flag '--help' to read about a specific subcommand.", strings.Join(args, " "))
	}
	if len(parsedArgs) == 0 {
		return nil
	}

	results := common.ExecuteConfigCommand(parsedArgs, common.ActionCreate)
	log.Infof("results: %+v", results)

	if results.FileInvalid {
		return fmt.Errorf("Failed to execute command: %v", results.Err)
	} else if results.NumHandled == 0 {
		if results.NumResources == 0 && parsedArgs["--dry-run"] == true {
			fmt.Println("No syntax problems, file is ready to be applied")
		} else if results.NumResources == 0 {
			return fmt.Errorf("No resources specified in file")
		} else if results.NumResources == 1 {
			return fmt.Errorf("Failed to create '%s' resource: %v", results.SingleKind, results.ResErrs)
		} else if results.SingleKind != "" {
			return fmt.Errorf("Failed to create any '%s' resources: %v", results.SingleKind, results.ResErrs)
		} else {
			return fmt.Errorf("Failed to create any resources: %v", results.ResErrs)
		}
	} else if len(results.ResErrs) == 0 {
		if results.SingleKind != "" {
			fmt.Printf("Successfully created %d '%s' resource(s)\n", results.NumHandled, results.SingleKind)
		} else {
			fmt.Printf("Successfully created %d resource(s)\n", results.NumHandled)
		}
	} else {
		if results.NumHandled-len(results.ResErrs) > 0 {
			fmt.Printf("Partial success: ")
			if results.SingleKind != "" {
				fmt.Printf("created the first %d out of %d '%s' resources:\n",
					results.NumHandled, results.NumResources, results.SingleKind)
			} else {
				fmt.Printf("created the first %d out of %d resources:\n",
					results.NumHandled, results.NumResources)
			}
		}
		return fmt.Errorf("Hit error: %v", results.ResErrs)
	}

	return nil
}
