// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/account"
	"github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/service/account/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_account_alternate_contact")
func ResourceAlternateContact() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAlternateContactCreate,
		ReadWithoutTimeout:   resourceAlternateContactRead,
		UpdateWithoutTimeout: resourceAlternateContactUpdate,
		DeleteWithoutTimeout: resourceAlternateContactDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"alternate_contact_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.AlternateContactType](),
			},
			"email_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[\w+=,.-]+@[\w.-]+\.[\w]+`), "must be a valid email address"),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"phone_number": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\s0-9()+-]+$`), "must be a valid phone number"),
			},
			"title": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
		},
	}
}

func resourceAlternateContactCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountClient()

	accountID := d.Get("account_id").(string)
	contactType := d.Get("alternate_contact_type").(string)
	id := AlternateContactCreateResourceID(accountID, contactType)
	input := &account.PutAlternateContactInput{
		AlternateContactType: types.AlternateContactType(contactType),
		EmailAddress:         aws.String(d.Get("email_address").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		PhoneNumber:          aws.String(d.Get("phone_number").(string)),
		Title:                aws.String(d.Get("title").(string)),
	}

	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	_, err := conn.PutAlternateContact(ctx, input)

	if err != nil {
		return diag.Errorf("creating Account Alternate Contact (%s): %s", id, err)
	}

	d.SetId(id)

	const (
		inARow = 2
	)
	_, err = retry.Operation(func(ctx context.Context) (*types.AlternateContact, error) {
		return FindAlternateContactByTwoPartKey(ctx, conn, accountID, contactType)
	}).UntilFoundN(inARow).Run(ctx, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("waiting for Account Alternate Contact (%s) create: %s", d.Id(), err)
	}

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountClient()

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindAlternateContactByTwoPartKey(ctx, conn, accountID, contactType)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Account Alternate Contact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Account Alternate Contact (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("alternate_contact_type", output.AlternateContactType)
	d.Set("email_address", output.EmailAddress)
	d.Set("name", output.Name)
	d.Set("phone_number", output.PhoneNumber)
	d.Set("title", output.Title)

	return nil
}

func resourceAlternateContactUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountClient()

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	email := d.Get("email_address").(string)
	name := d.Get("name").(string)
	phone := d.Get("phone_number").(string)
	title := d.Get("title").(string)

	input := &account.PutAlternateContactInput{
		AlternateContactType: types.AlternateContactType(contactType),
		EmailAddress:         aws.String(email),
		Name:                 aws.String(name),
		PhoneNumber:          aws.String(phone),
		Title:                aws.String(title),
	}

	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	_, err = conn.PutAlternateContact(ctx, input)

	if err != nil {
		return diag.Errorf("updating Account Alternate Contact (%s): %s", d.Id(), err)
	}

	_, err = retry.Operation(func(ctx context.Context) (*types.AlternateContact, error) {
		return FindAlternateContactByTwoPartKey(ctx, conn, accountID, contactType)
	}).If(func(v *types.AlternateContact, err error) (bool, error) {
		if err != nil {
			return false, err
		}

		equal := email == aws.ToString(v.EmailAddress) && name == aws.ToString(v.Name) && phone == aws.ToString(v.PhoneNumber) && title == aws.ToString(v.Title)

		return !equal, nil
	}).Run(ctx, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return diag.Errorf("waiting for Account Alternate Contact (%s) update: %s", d.Id(), err)
	}

	return resourceAlternateContactRead(ctx, d, meta)
}

func resourceAlternateContactDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccountClient()

	accountID, contactType, err := AlternateContactParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &account.DeleteAlternateContactInput{
		AlternateContactType: types.AlternateContactType(contactType),
	}
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	log.Printf("[DEBUG] Deleting Account Alternate Contact: %s", d.Id())
	_, err = conn.DeleteAlternateContact(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Account Alternate Contact (%s): %s", d.Id(), err)
	}

	_, err = retry.Operation(func(ctx context.Context) (*types.AlternateContact, error) {
		return FindAlternateContactByTwoPartKey(ctx, conn, accountID, contactType)
	}).UntilNotFound().Run(ctx, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("waiting for Account Alternate Contact (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindAlternateContactByTwoPartKey(ctx context.Context, conn *account.Client, accountID, contactType string) (*types.AlternateContact, error) {
	input := &account.GetAlternateContactInput{
		AlternateContactType: types.AlternateContactType(contactType),
	}
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetAlternateContact(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlternateContact == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AlternateContact, nil
}

const alternateContactResourceIDSeparator = "/"

func AlternateContactCreateResourceID(accountID, contactType string) string {
	parts := []string{accountID, contactType}
	id := strings.Join(parts, alternateContactResourceIDSeparator)

	return id
}

func AlternateContactParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, alternateContactResourceIDSeparator)

	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case 2:
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected ContactType or AccountID%[2]sContactType", id, alternateContactResourceIDSeparator)
	}
}
