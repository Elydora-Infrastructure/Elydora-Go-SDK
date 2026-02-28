package elydora

import "fmt"

// RegisterAgent registers a new AI agent within the organization.
func (c *Client) RegisterAgent(req *RegisterAgentRequest) (*RegisterAgentResponse, error) {
	var result RegisterAgentResponse
	if err := c.doPost("/agents", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAgent retrieves an agent by ID.
func (c *Client) GetAgent(agentID string) (*GetAgentResponse, error) {
	var result GetAgentResponse
	if err := c.doGet(fmt.Sprintf("/agents/%s", agentID), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FreezeAgent freezes an agent, preventing it from submitting operations.
func (c *Client) FreezeAgent(agentID, reason string) error {
	return c.doPost(fmt.Sprintf("/agents/%s/freeze", agentID), &FreezeAgentRequest{
		Reason: reason,
	}, nil)
}

// RevokeKey revokes a specific key for an agent.
func (c *Client) RevokeKey(agentID, kid, reason string) error {
	return c.doPost(fmt.Sprintf("/agents/%s/revoke-key", agentID), &RevokeAgentRequest{
		KID:    kid,
		Reason: reason,
	}, nil)
}
