// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -UpdateTags
//go:generate go run ../../generate/listpages/main.go -ListOps=ListLicenseConfigurations,ListLicenseSpecificationsForResource,ListReceivedLicenses,ListDistributedGrants
// ONLY generate directives and package declaration! Do not add anything else to this file.

package licensemanager
