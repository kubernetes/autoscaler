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

// Statement is an empty AST mostly used for transitioning states.
func newStatement() AST {
	return newAST(ASTKindStatement, AST{})
}

// SectionStatement represents a section AST
func newSectionStatement(tok Token) AST {
	return newASTWithRootToken(ASTKindSectionStatement, tok)
}

// ExprStatement represents a completed expression AST
func newExprStatement(ast AST) AST {
	return newAST(ASTKindExprStatement, ast)
}

// CommentStatement represents a comment in the ini definition.
//
//	grammar:
//	comment -> #comment' | ;comment'
//	comment' -> epsilon | value
func newCommentStatement(tok Token) AST {
	return newAST(ASTKindCommentStatement, newExpression(tok))
}

// CompletedSectionStatement represents a completed section
func newCompletedSectionStatement(ast AST) AST {
	return newAST(ASTKindCompletedSectionStatement, ast)
}

// SkipStatement is used to skip whole statements
func newSkipStatement(ast AST) AST {
	return newAST(ASTKindSkipStatement, ast)
}
