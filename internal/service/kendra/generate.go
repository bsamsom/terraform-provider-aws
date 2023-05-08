// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -TagInIDElem=ResourceARN -ListTags -ListTagsInIDElem=ResourceARN -ServiceTagsSlice -UpdateTags -UntagInTagsElem=TagKeys
// ONLY generate directives and package declaration! Do not add anything else to this file.

package kendra
