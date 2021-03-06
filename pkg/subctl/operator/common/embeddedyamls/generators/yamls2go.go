/*
© 2019 Red Hat, Inc. and others.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

var files = []string{
	"deploy/crds/submariner.io_submariners.yaml",
	"deploy/crds/submariner.io_servicediscoveries.yaml",
	"deploy/submariner/crds/submariner.io_clusters.yaml",
	"deploy/submariner/crds/submariner.io_endpoints.yaml",
	"deploy/submariner/crds/submariner.io_gateways.yaml",
	"deploy/lighthouse/crds/lighthouse.submariner.io_multiclusterservices.yaml",
	"deploy/lighthouse/crds/lighthouse.submariner.io_serviceexports.yaml",
	"deploy/lighthouse/crds/lighthouse.submariner.io_serviceimports.yaml",
	"deploy/mcsapi/crds/multicluster.x_k8s.io_serviceexports.yaml",
	"deploy/mcsapi/crds/multicluster.x_k8s.io_serviceimports.yaml",
	"config/rbac/cluster_role.yaml",
	"config/rbac/cluster_role_binding.yaml",
	"config/rbac/globalnet_cluster_role.yaml",
	"config/rbac/globalnet_cluster_role_binding.yaml",
	"config/rbac/lighthouse_cluster_role_binding.yaml",
	"config/rbac/lighthouse_cluster_role.yaml",
	"config/rbac/role.yaml",
	"config/rbac/role_binding.yaml",
	"config/rbac/service_account.yaml",
}

// Reads all .yaml files in the crdDirectory
// and encodes them as constants in crdyamls.go
func main() {
	if len(os.Args) < 3 {
		fmt.Println("yamls2go needs two arguments, the base directory containing the YAML files, and the target directory")
		os.Exit(1)
	}

	yamlsDirectory := os.Args[1]
	goDirectory := os.Args[2]

	fmt.Println("Generating yamls.go")
	out, err := os.Create(goDirectory + string(os.PathSeparator) + "yamls.go")
	panicOnErr(err)

	_, err = out.WriteString("// This file is auto-generated by yamls2go.go\n" +
		"package embeddedyamls\n\nconst (\n")
	panicOnErr(err)

	// Raw string literals can’t contain backticks (which enclose the literals)
	// and there’s no way to escape them. Some YAML files we need to embed include
	// backticks... To work around this, without having to deal with all the
	// subtleties of wrapping arbitrary YAML in interpreted string literals, we
	// split raw string literals when we encounter backticks in the source YAML,
	// and add the backtick-enclosed string as an interpreted string:
	//
	// `resourceLock:
	//    description: The type of resource object that is used for locking
	//      during leader election. Supported options are ` + "`configmaps`" + ` (default)
	//      and ` + "`endpoints`" + `.
	//    type: string`

	re := regexp.MustCompile("`([^`]*)`")

	for _, f := range files {
		_, err = out.WriteString("\t" + constName(f) + " = `")
		panicOnErr(err)

		fmt.Println(f)
		contents, err := ioutil.ReadFile(path.Join(yamlsDirectory, f))
		panicOnErr(err)

		_, err = out.Write(re.ReplaceAll(contents, []byte("` + \"`$1`\" + `")))
		panicOnErr(err)

		_, err = out.WriteString("`\n")
		panicOnErr(err)
	}
	_, err = out.WriteString(")\n")
	panicOnErr(err)

	err = out.Close()
	panicOnErr(err)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func constName(filename string) string {
	return strings.Title(strings.ReplaceAll(
		strings.ReplaceAll(filename,
			".", "_"),
		"/", "_"))
}
