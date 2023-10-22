package provider

import (
	"context"
	"encoding/base64"

	"github.com/fhke/terraform-provider-jks/jks"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewKeystoreDataSource() datasource.DataSource {
	return &KeystoreDataSource{}
}

// KeystoreDataSource defines the data source implementation.
type KeystoreDataSource struct{}

// KeystoreDataSourceModel describes the data source data model.
type KeystoreDataSourceModel struct {
	// Input values
	KeyPair  types.Set    `tfsdk:"key_pair"`
	Password types.String `tfsdk:"password"`
	// Computed values
	JksB64 types.String `tfsdk:"jks_base64"`
}

func (d *KeystoreDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keystore"
}

func (d *KeystoreDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a JKS keystore",

		Attributes: map[string]schema.Attribute{
			"password": schema.StringAttribute{
				Description: "Password for keystore.",
				Required:    true,
			},
			"jks_base64": schema.StringAttribute{
				Description: "Base 64 encoded keystore, in JKS format",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"key_pair": schema.SetNestedBlock{
				Description: "Block defining a cert & key pair.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alias": schema.StringAttribute{
							Required:    true,
							Description: "Alias for key pair. Must be unique within keystore.",
						},
						"certificate": schema.StringAttribute{
							Required:    true,
							Description: "Certificate in PEM format.",
						},
						"private_key": schema.StringAttribute{
							Required:    true,
							Description: "Private key for certificate in PEM format.",
						},
						"intermediate_certificates": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "List of intermediate certificate authority certificates in PEM format. Root certificates should not be added here.",
						},
					},
				},
			},
		},
	}
}

func (d *KeystoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KeystoreDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// create jks builder
	bld := jks.NewKeystoreBuilder()

	// set store password
	bld.SetPassword(data.Password.ValueString())

	for _, kpElem := range data.KeyPair.Elements() {
		keyPair := kpElem.(types.Object).Attributes()

		// get intermediate certs as [][]byte
		caCerts := make([][]byte, 0)
		for _, crtElem := range keyPair["intermediate_certificates"].(types.List).Elements() {
			caCerts = append(caCerts, []byte(crtElem.(types.String).ValueString()))
		}

		// Add cert to store
		bld.AddCert(
			keyPair["alias"].(types.String).ValueString(),
			[]byte(keyPair["certificate"].(types.String).ValueString()),
			[]byte(keyPair["private_key"].(types.String).ValueString()),
			caCerts...,
		)
	}

	// build jks keystore
	jksData, err := bld.Build()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating JKS keystore",
			err.Error(),
		)
		return
	}

	// base64 encode jks & add to model
	data.JksB64 = types.StringValue(base64.StdEncoding.EncodeToString(jksData))

	// save model
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
