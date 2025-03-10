// Copyright 2022 Woodpecker Authors
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

package constant

var PrivilegedPlugins = []string{
	"plugins/docker",
	"plugins/gcr",
	"plugins/ecr",
	"woodpeckerci/plugin-docker",
	"woodpeckerci/plugin-docker-buildx",
}

// DefaultConfigOrder represent the priority in witch woodpecker search for a pipeline config by default
// folders are indicated by supplying a trailing /
var DefaultConfigOrder = [...]string{
	".woodpecker/",
	".woodpecker.yml",
	".woodpecker.yaml",
	".drone.yml",
}

const (
	DefaultCloneImage = "docker.io/woodpeckerci/plugin-git:v1.6.0"
)
