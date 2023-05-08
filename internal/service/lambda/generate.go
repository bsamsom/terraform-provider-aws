// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsMap -TagInIDElem=Resource -UpdateTags -AWSSDKVersion=2 -KVTValues -SkipTypesImp
// ONLY generate directives and package declaration! Do not add anything else to this file.

package lambda
