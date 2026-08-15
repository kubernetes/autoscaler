/*
Copyright 2023 The Kubernetes Authors.

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

package ini

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

// Walk will traverse the AST using the v, the Visitor.
func Walk(tree []AST, v Visitor) error {
	for _, node := range tree {
		switch node.Kind {
		case ASTKindExpr,
			ASTKindExprStatement:

			if err := v.VisitExpr(node); err != nil {
				return err
			}
		case ASTKindStatement,
			ASTKindCompletedSectionStatement,
			ASTKindNestedSectionStatement,
			ASTKindCompletedNestedSectionStatement:

			if err := v.VisitStatement(node); err != nil {
				return err
			}
		}
	}

	return nil
}
