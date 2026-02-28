package elydora

import "fmt"

// CreateExport creates a new compliance export job.
func (c *Client) CreateExport(params *CreateExportRequest) (*CreateExportResponse, error) {
	var result CreateExportResponse
	if err := c.doPost("/exports", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListExports lists all exports for the organization.
func (c *Client) ListExports() (*ListExportsResponse, error) {
	var result ListExportsResponse
	if err := c.doGet("/exports", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetExport retrieves a specific export by ID.
func (c *Client) GetExport(exportID string) (*GetExportResponse, error) {
	var result GetExportResponse
	if err := c.doGet(fmt.Sprintf("/exports/%s", exportID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
