// Copyright © 2017 The Kubicorn Authors
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

package auth

import (
	"io/ioutil"
	"golang.org/x/crypto/ssh"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh/agent"
	"net"
	"os"
)

// SystemAgent returns system agent if it exists.
func SystemAgent() agent.Agent {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return agent.NewClient(sshAgent)
	}
	return nil
}

// CheckKey checks is key present in the agent.
func CheckKey(agent agent.Agent, pubkey string) error {
	p, err := ioutil.ReadFile(pubkey)
	if err != nil {
		return err
	}

	authkey, _, _, _, _ := ssh.ParseAuthorizedKey(p)
	if err != nil {
		return err
	}
	parsedkey := authkey.Marshal()

	list, err := agent.List()
	if err != nil {
		return err
	}

	for _, key := range list {
		if bytes.Equal(key.Blob, parsedkey) {
			return nil
		}
	}

	return fmt.Errorf("key not found in keyring")
}