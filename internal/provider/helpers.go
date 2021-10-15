package provider

import (
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func returnResponseErr(resp *http.Response, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	diags = append(diags, diag.FromErr(err)...)

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	diags = append(diags, diag.Errorf(string(b))...)
	return diags

}
