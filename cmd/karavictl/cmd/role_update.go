// Copyright © 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
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

package cmd

import (
	"errors"
	"fmt"
	"karavi-authorization/internal/roles"
	"strings"

	"github.com/spf13/cobra"
)

// NewRoleUpdateCmd creates a new update command
func NewRoleUpdateCmd() *cobra.Command {
	roleUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update one or more Karavi roles",
		Long:  `Updates one or more Karavi roles`,
		Run: func(cmd *cobra.Command, args []string) {
			outFormat := "failed to update role: %+v\n"

			roleFlags, err := cmd.Flags().GetStringSlice("role")
			if err != nil {
				reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf(outFormat, err))
			}

			var rff roles.JSON
			if len(roleFlags) == 0 {
				reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf(outFormat, errors.New("no input")))
			}

			for _, v := range roleFlags {
				t := strings.Split(v, "=")
				err = rff.Add(roles.NewInstance(t[0], t[1:]...))
				if err != nil {
					reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf(outFormat, err))
				}
			}

			existingRoles, err := GetRoles()
			if err != nil {
				reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf(outFormat, err))
			}

			for _, rls := range rff.Instances() {
				if existingRoles.Get(rls.RoleKey) == nil {
					reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf("%s role does not exist. Try create command", rls.Name))
				}

				err = validateRole(rls)
				if err != nil {
					reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf("%s failed validation: %+v", rls.Name, err))
				}

				err = existingRoles.Remove(rls)
				if err != nil {
					reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf("%s failed to update: %+v", rls.Name, err))
				}
				err := existingRoles.Add(rls)
				if err != nil {
					reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf("adding %s failed: %+v", rls.Name, err))
				}
			}
			if err = modifyCommonConfigMap(existingRoles); err != nil {
				reportErrorAndExit(JSONOutput, cmd.ErrOrStderr(), fmt.Errorf(outFormat, err))
			}
		},
	}
	roleUpdateCmd.Flags().StringSlice("role", []string{}, "role in the form <name>=<type>=<id>=<pool>=<quota>")
	return roleUpdateCmd
}
